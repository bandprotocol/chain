package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SetLatestPrices sets the latest prices in the store
func (k Keeper) SetLatestPrices(ctx sdk.Context, latestPrices types.LatestPrices) {
	ctx.KVStore(k.storeKey).
		Set(types.LatestPricesStoreKey(latestPrices.TunnelID), k.cdc.MustMarshal(&latestPrices))
}

// GetLatestPrices gets the latest prices from the store
func (k Keeper) GetLatestPrices(ctx sdk.Context, tunnelID uint64) (types.LatestPrices, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.LatestPricesStoreKey(tunnelID))
	if bz == nil {
		return types.LatestPrices{}, types.ErrLatestPricesNotFound.Wrapf("tunnelID: %d", tunnelID)
	}

	var latestPrices types.LatestPrices
	k.cdc.MustUnmarshal(bz, &latestPrices)
	return latestPrices, nil
}

// MustGetLatestPrices retrieves the latest prices by its tunnel ID. Panics if the prices does not exist.
func (k Keeper) MustGetLatestPrices(ctx sdk.Context, tunnelID uint64) types.LatestPrices {
	latestPrices, err := k.GetLatestPrices(ctx, tunnelID)
	if err != nil {
		panic(err)
	}
	return latestPrices
}

// GetAllLatestPrices gets all the latest prices from the store
func (k Keeper) GetAllLatestPrices(ctx sdk.Context) []types.LatestPrices {
	var allLatestPrices []types.LatestPrices
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.LatestPricesStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var latestPrices types.LatestPrices
		k.cdc.MustUnmarshal(iterator.Value(), &latestPrices)
		allLatestPrices = append(allLatestPrices, latestPrices)
	}
	return allLatestPrices
}
