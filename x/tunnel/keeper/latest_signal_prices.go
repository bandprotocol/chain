package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SetLatestSignalPrices sets the latest signal prices in the store
func (k Keeper) SetLatestSignalPrices(ctx sdk.Context, latestSignalPrices types.LatestSignalPrices) {
	ctx.KVStore(k.storeKey).
		Set(types.LatestSignalPricesStoreKey(latestSignalPrices.TunnelID), k.cdc.MustMarshal(&latestSignalPrices))
}

// GetLatestSignalPrices gets the latest signal prices from the store
func (k Keeper) GetLatestSignalPrices(ctx sdk.Context, tunnelID uint64) (types.LatestSignalPrices, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.LatestSignalPricesStoreKey(tunnelID))
	if bz == nil {
		return types.LatestSignalPrices{}, types.ErrLatestSignalPricesNotFound.Wrapf("tunnelID: %d", tunnelID)
	}

	var latestSignalPrices types.LatestSignalPrices
	k.cdc.MustUnmarshal(bz, &latestSignalPrices)
	return latestSignalPrices, nil
}

// MustGetLatestSignalPrices retrieves the latest signal prices by its tunnel ID. Panics if the signal prices does not exist.
func (k Keeper) MustGetLatestSignalPrices(ctx sdk.Context, tunnelID uint64) types.LatestSignalPrices {
	latestSignalPrices, err := k.GetLatestSignalPrices(ctx, tunnelID)
	if err != nil {
		panic(err)
	}
	return latestSignalPrices
}

// GetAllLatestSignalPrices gets all the latest signal prices from the store
func (k Keeper) GetAllLatestSignalPrices(ctx sdk.Context) []types.LatestSignalPrices {
	var allLatestSignalPrices []types.LatestSignalPrices
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.LatestSignalPricesStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var latestSignalPrices types.LatestSignalPrices
		k.cdc.MustUnmarshal(iterator.Value(), &latestSignalPrices)
		allLatestSignalPrices = append(allLatestSignalPrices, latestSignalPrices)
	}
	return allLatestSignalPrices
}
