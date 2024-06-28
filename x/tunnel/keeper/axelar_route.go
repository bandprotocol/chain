package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (k Keeper) SetAxelarRouteCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.AxelarRouteCountStoreKey, sdk.Uint64ToBigEndian(count))
}

func (k Keeper) GetAxelarRouteCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.AxelarRouteCountStoreKey))
}

func (k Keeper) GetNextAxelarRouteID(ctx sdk.Context) uint64 {
	routeNumber := k.GetAxelarRouteCount(ctx) + 1
	k.SetAxelarRouteCount(ctx, routeNumber)
	return routeNumber
}
