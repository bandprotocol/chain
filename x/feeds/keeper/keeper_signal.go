package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (k Keeper) GetDelegatorSignals(ctx sdk.Context, delegator sdk.AccAddress) []types.Signal {
	bz := ctx.KVStore(k.storeKey).Get(types.DelegatorSignalStoreKey(delegator))
	if bz == nil {
		return nil
	}

	var s types.DelegatorSignals
	k.cdc.MustUnmarshal(bz, &s)

	return s.Signals
}

func (k Keeper) DeleteDelegatorSignals(ctx sdk.Context, delegator sdk.AccAddress) {
	ctx.KVStore(k.storeKey).
		Delete(types.DelegatorSignalStoreKey(delegator))
}

func (k Keeper) SetDelegatorSignals(ctx sdk.Context, signals types.DelegatorSignals) {
	ctx.KVStore(k.storeKey).
		Set(types.DelegatorSignalStoreKey(sdk.MustAccAddressFromBech32(signals.Delegator)), k.cdc.MustMarshal(&signals))
}

func (k Keeper) GetDelegatorSignalsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DelegatorSignalStoreKeyPrefix)
}

func (k Keeper) GetAllDelegatorSignals(ctx sdk.Context) (delegatorSignalsList []types.DelegatorSignals) {
	iterator := k.GetDelegatorSignalsIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var ds types.DelegatorSignals
		k.cdc.MustUnmarshal(iterator.Value(), &ds)
		delegatorSignalsList = append(delegatorSignalsList, ds)
	}

	return delegatorSignalsList
}

func (k Keeper) SetAllDelegatorSignals(ctx sdk.Context, delegatorSignalsList []types.DelegatorSignals) {
	for _, ds := range delegatorSignalsList {
		k.SetDelegatorSignals(ctx, ds)
	}
}
