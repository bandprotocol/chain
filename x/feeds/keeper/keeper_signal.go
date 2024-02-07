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

	var s types.Signals
	k.cdc.MustUnmarshal(bz, &s)

	return s.Signals
}

func (k Keeper) SetDelegatorSignals(ctx sdk.Context, delegator sdk.AccAddress, signals []types.Signal) {
	ctx.KVStore(k.storeKey).
		Set(types.DelegatorSignalStoreKey(delegator), k.cdc.MustMarshal(&types.Signals{Signals: signals}))
}
