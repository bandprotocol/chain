package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (k Keeper) SetTSSRouteCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.TSSRouteCountStoreKey, sdk.Uint64ToBigEndian(count))
}

func (k Keeper) GetTSSRouteCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.TSSRouteCountStoreKey))
}

func (k Keeper) GetNextTSSRouteID(ctx sdk.Context) uint64 {
	routeNumber := k.GetTSSRouteCount(ctx) + 1
	k.SetTSSRouteCount(ctx, routeNumber)
	return routeNumber
}
