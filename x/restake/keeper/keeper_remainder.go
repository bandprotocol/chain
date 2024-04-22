package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetRemainder(ctx sdk.Context) types.Remainder {
	bz := ctx.KVStore(k.storeKey).Get(types.RemainderStoreKey)
	if bz == nil {
		panic("Stored remainder should not have been nil")
	}

	var remainder types.Remainder
	k.cdc.MustUnmarshal(bz, &remainder)

	return remainder
}

func (k Keeper) SetRemainder(ctx sdk.Context, remainder types.Remainder) {
	ctx.KVStore(k.storeKey).Set(types.RemainderStoreKey, k.cdc.MustMarshal(&remainder))
}

func (k Keeper) ProcessRemainder(ctx sdk.Context) {
	remainder := k.GetRemainder(ctx)
	truncatedCoins, changedCoins := remainder.Amounts.TruncateDecimal()

	if !truncatedCoins.IsZero() {
		address := k.authKeeper.GetModuleAddress(types.ModuleName)
		err := k.distrKeeper.FundCommunityPool(ctx, truncatedCoins, address)
		if err != nil {
			return
		}

		remainder.Amounts = changedCoins
		k.SetRemainder(ctx, remainder)
	}
}

func (k Keeper) addRemainderAmount(ctx sdk.Context, decCoins sdk.DecCoins) {
	remainder := k.GetRemainder(ctx)
	remainder.Amounts = remainder.Amounts.Add(decCoins...)
	k.SetRemainder(ctx, remainder)
}
