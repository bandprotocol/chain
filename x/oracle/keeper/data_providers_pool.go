package oraclekeeper

import (
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) FundOraclePool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
		return err
	}

	oraclePool := k.GetOraclePool(ctx)
	oraclePool.DataProvidersPool = oraclePool.DataProvidersPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	k.SetOraclePool(ctx, oraclePool)

	return nil
}

func (k Keeper) WithdrawOraclePool(ctx sdk.Context, amount sdk.Coins, recipient sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, amount); err != nil {
		return err
	}

	oraclePool := k.GetOraclePool(ctx)
	diff, hasNeg := oraclePool.DataProvidersPool.SafeSub(sdk.NewDecCoinsFromCoins(amount...))
	if hasNeg {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "data providers pool does not have enough funds")
	}
	oraclePool.DataProvidersPool = diff
	k.SetOraclePool(ctx, oraclePool)
	return nil
}
