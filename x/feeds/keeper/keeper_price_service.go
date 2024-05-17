package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetPriceService gets the current price service.
func (k Keeper) GetPriceService(ctx sdk.Context) (ps types.PriceService) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceServiceStoreKey)
	if bz == nil {
		return ps
	}

	k.cdc.MustUnmarshal(bz, &ps)

	return ps
}

// SetPriceService sets new price service to the store.
func (k Keeper) SetPriceService(ctx sdk.Context, ps types.PriceService) error {
	if err := ps.Validate(); err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.PriceServiceStoreKey, k.cdc.MustMarshal(&ps))

	return nil
}
