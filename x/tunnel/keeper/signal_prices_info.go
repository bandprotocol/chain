package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetSignalPricesInfo sets the signal prices info in the store
func (k Keeper) SetSignalPricesInfo(ctx sdk.Context, signalPricesInfo types.SignalPricesInfo) {
	ctx.KVStore(k.storeKey).
		Set(types.SignalPricesInfoStoreKey(signalPricesInfo.TunnelID), k.cdc.MustMarshal(&signalPricesInfo))
}

// GetSignalPricesInfo gets the signal prices info from the store
func (k Keeper) GetSignalPricesInfo(ctx sdk.Context, tunnelID uint64) (types.SignalPricesInfo, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SignalPricesInfoStoreKey(tunnelID))
	if bz == nil {
		return types.SignalPricesInfo{}, types.ErrSignalPricesInfoNotFound.Wrapf("tunnelID: %d", tunnelID)
	}

	var signalPricesInfo types.SignalPricesInfo
	k.cdc.MustUnmarshal(bz, &signalPricesInfo)
	return signalPricesInfo, nil
}

// MustGetSignalPricesInfo gets the signal prices info from the store and panics if it does not exist
func (k Keeper) MustGetSignalPricesInfo(ctx sdk.Context, tunnelID uint64) types.SignalPricesInfo {
	signalPricesInfo, err := k.GetSignalPricesInfo(ctx, tunnelID)
	if err != nil {
		panic(err)
	}
	return signalPricesInfo
}

// GetSignalPricesInfos gets all signal prices infos from the store
func (k Keeper) GetSignalPricesInfos(ctx sdk.Context) []types.SignalPricesInfo {
	var signalPricesInfos []types.SignalPricesInfo
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SignalPricesInfoStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var signalPricesInfo types.SignalPricesInfo
		k.cdc.MustUnmarshal(iterator.Value(), &signalPricesInfo)
		signalPricesInfos = append(signalPricesInfos, signalPricesInfo)
	}
	return signalPricesInfos
}
