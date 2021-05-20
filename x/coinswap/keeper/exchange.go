package keeper

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"strings"
)

// ExchangeDenom exchanges given amount
func (k Keeper) ExchangeDenom(ctx sdk.Context, fromDenom, toDenom string, amt sdk.Coin, requester sdk.AccAddress) error {
	pair, err := k.GetExchangePair(ctx, fromDenom, toDenom)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to get exchange rate")
	}
	convertedAmt, err := k.Convert(ctx, amt, pair)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to convert coins")
	}
	return k.Exchange(ctx, amt, convertedAmt, requester)
}

// Exchange withdraws coins to community pool
func (k Keeper) Exchange(ctx sdk.Context, initialAmt sdk.Coin, convertedAmt sdk.Coin, requester sdk.AccAddress) error {
	// first send source tokens to module
	if err := k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(initialAmt), requester); err != nil {
		return sdkerrors.Wrapf(err, "sending coins from account: %s, to module: %s", requester.String(), distrtypes.ModuleName)
	}
	if err := k.oracleKeeper.WithdrawOraclePool(ctx, sdk.NewCoins(convertedAmt), requester); err != nil {
		return sdkerrors.Wrapf(err, "sending coins from module: %s, to account: %s", oracletypes.ModuleName, requester.String())
	}
	return nil
}

// Convert converts coins with the given exchange rate
func (k Keeper) Convert(ctx sdk.Context, amt sdk.Coin, pair coinswaptypes.Exchange) (sdk.Coin, error) {
	rate := k.CalculateRate(ctx, pair.RateMultiplier)
	if rate.GT(amt.Amount.ToDec()) {
		return sdk.Coin{}, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "current rate: %s is higher then amount provided: %s", rate.String(), amt.String())
	}
	convertedAmt := sdk.NewDecCoinFromDec(pair.To, amt.Amount.ToDec().QuoRoundUp(rate))
	base, remainder := convertedAmt.TruncateDecimal()
	if !remainder.IsZero() {
		k.Logger(ctx).With("coins", remainder.String()).Info("performing exchange according to limited precision some coins are lost")
	}
	return base, nil
}

// GetRate returns the exchange rate for the given pair
func (k Keeper) GetRate(ctx sdk.Context, fromDenom, toDenom string) (sdk.Dec, error) {
	pair, err := k.GetExchangePair(ctx, fromDenom, toDenom)
	if err != nil {
		return sdk.Dec{}, sdkerrors.Wrapf(err, "failed to get rate from: %s, to: %s", fromDenom, toDenom)
	}
	return k.CalculateRate(ctx, pair.RateMultiplier), nil
}

// CalculateRate calculates the exchange rate for the given rate multiplier
func (k Keeper) CalculateRate(ctx sdk.Context, rateMultiplier sdk.Dec) sdk.Dec {
	initialRate := k.GetInitialRate(ctx)
	return initialRate.Mul(rateMultiplier)
}

// GetExchangePair returns rate multiplier for the given denoms
func (k Keeper) GetExchangePair(ctx sdk.Context, fromDenom, toDenom string) (coinswaptypes.Exchange, error) {
	params := k.GetParams(ctx)
	for _, ex := range params.Exchanges {
		if strings.ToLower(ex.From) == strings.ToLower(fromDenom) && strings.ToLower(ex.To) == strings.ToLower(toDenom) {
			return ex, nil
		}
	}
	return coinswaptypes.Exchange{}, sdkerrors.Wrapf(coinswaptypes.ErrInvalidExchangeDenom, "failed to find pair %s:%s", fromDenom, toDenom)
}
