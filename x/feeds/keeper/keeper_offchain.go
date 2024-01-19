package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (k Keeper) GetOffChain(ctx sdk.Context) (types.OffChain, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.OffChainStoreKey)
	if bz == nil {
		return types.OffChain{}, types.ErrPriceNotFound.Wrap("failed to get off-chain detail")
	}

	var oc types.OffChain
	k.cdc.MustUnmarshal(bz, &oc)

	return oc, nil
}

func (k Keeper) SetOffChain(ctx sdk.Context, offChain types.OffChain) error {
	if err := offChain.Validate(); err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.OffChainStoreKey, k.cdc.MustMarshal(&offChain))
	return nil
}
