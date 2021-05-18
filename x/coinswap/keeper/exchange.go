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

	return k.Exchange(ctx, amt, pair, requester)
}

func (k Keeper) Exchange(ctx sdk.Context, amt sdk.Coin, pair coinswaptypes.Exchange, requester sdk.AccAddress) error {
	rate, err := k.CalculateRate(ctx, pair.RateMultiplier)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to calculate the exchange rate")
	}
	if rate.GT(amt.Amount.ToDec()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "current rate: %s is higher then amount provided: %s", rate.String(), amt.String())
	}

	convertedAmt := sdk.NewDecCoinFromDec(pair.To, amt.Amount.ToDec().QuoRoundUp(rate))

	// first send source tokens to module
	if err := k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), requester); err != nil {
		return sdkerrors.Wrapf(err, "sending coins from account: %s, to module: %s", requester.String(), distrtypes.ModuleName)
	}

	toSend, remainder := convertedAmt.TruncateDecimal()
	if !remainder.IsZero() {
		k.Logger(ctx).With("coins", remainder.String()).Info("performing exchange according to limited precision some coins are lost")
	}

	if err := k.oracleKeeper.WithdrawOraclePool(ctx, sdk.NewCoins(toSend), requester); err != nil {
		return sdkerrors.Wrapf(err, "sending coins from module: %s, to account: %s", oracletypes.ModuleName, requester.String())
	}

	return nil
}

// GetRate returns the exchange rate for the given pair
func (k Keeper) GetRate(ctx sdk.Context, fromDenom, toDenom string) (sdk.Dec, error) {
	pair, err := k.GetExchangePair(ctx, fromDenom, toDenom)
	if err != nil {
		return sdk.Dec{}, sdkerrors.Wrapf(err, "failed to get rate from: %s, to: %s", fromDenom, toDenom)
	}
	return k.CalculateRate(ctx, pair.RateMultiplier)
}

// CalculateRate calculates the exchange rate for the given rate multiplier
func (k Keeper) CalculateRate(ctx sdk.Context, rateMultiplier sdk.Dec) (sdk.Dec, error) {
	initialRate := k.GetInitialRate(ctx)
	return initialRate.Mul(rateMultiplier), nil
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
