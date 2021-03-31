package keeper

import (
	"github.com/GeoDB-Limited/odincore/chain/x/coinswap/types"
	commontypes "github.com/GeoDB-Limited/odincore/chain/x/common/types"
	"github.com/GeoDB-Limited/odincore/chain/x/oracle"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

func (k Keeper) ExchangeDenom(ctx sdk.Context, from, to commontypes.Denom, amt sdk.Coin, requester sdk.AccAddress) error {

	// convert source amount to destination amount according to rate
	convertedAmt, err := k.convertToRate(ctx, from, to, amt)
	if err != nil {
		return sdkerrors.Wrap(err, "converting rate")
	}

	// first send source tokens to module
	err = k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(amt), requester)
	if err != nil {
		return sdkerrors.Wrapf(err, "sending coins from account: %s, to module: %s", requester.String(), distr.ModuleName)
	}

	toSend, remainder := convertedAmt.TruncateDecimal()
	if !remainder.IsZero() {
		k.Logger(ctx).With("coins", remainder.String()).Info("performing exchange according to limited precision some coins are lost")
	}

	err = k.oracleKeeper.WithdrawOraclePool(ctx, sdk.NewCoins(toSend), requester)
	if err != nil {
		return sdkerrors.Wrapf(err, "sending coins from module: %s, to account: %s", oracle.ModuleName, requester.String())
	}

	return nil
}

func (k Keeper) GetRate(ctx sdk.Context) sdk.Dec {
	params := types.Params{}
	k.paramSpace.GetParamSet(ctx, &params)
	initialRate := k.GetInitialRate(ctx)
	return initialRate.Mul(params.RateMultiplier)
}

// returns the converted amount according to current rate
func (k Keeper) convertToRate(ctx sdk.Context, from, to commontypes.Denom, amt sdk.Coin) (sdk.DecCoin, error) {
	rate := k.GetRate(ctx)
	if rate.GT(amt.Amount.ToDec()) {
		return sdk.DecCoin{}, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "current rate: %s is higher then amount provided: %s", rate.String(), amt.String())
	}
	convertedAmt := amt.Amount.ToDec().QuoRoundUp(rate)
	return sdk.NewDecCoinFromDec(to.String(), convertedAmt), nil
}
