package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (k Keeper) GetOffChain(ctx sdk.Context) (oc types.OffChain) {
	bz := ctx.KVStore(k.storeKey).Get(types.OffChainStoreKey)
	if bz == nil {
		return oc
	}

	k.cdc.MustUnmarshal(bz, &oc)

	return oc
}

func (k Keeper) SetOffChain(ctx sdk.Context, offChain types.OffChain) error {
	if err := offChain.Validate(); err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.OffChainStoreKey, k.cdc.MustMarshal(&offChain))
	return nil
}
