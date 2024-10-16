package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// GetCurrentFeeds gets the current supported feeds.
func (k Keeper) GetCurrentFeeds(ctx sdk.Context) (sp types.CurrentFeeds) {
	bz := ctx.KVStore(k.storeKey).Get(types.CurrentFeedsStoreKey)
	if bz == nil {
		return sp
	}

	k.cdc.MustUnmarshal(bz, &sp)

	return sp
}

// SetCurrentFeeds sets new supported feeds to the store.
func (k Keeper) SetCurrentFeeds(ctx sdk.Context, feeds []types.Feed) {
	cf := types.NewCurrentFeeds(feeds, ctx.BlockTime().Unix(), ctx.BlockHeight())

	ctx.KVStore(k.storeKey).Set(types.CurrentFeedsStoreKey, k.cdc.MustMarshal(&cf))
	emitEventUpdateCurrentFeeds(ctx, cf)
}

// CalculateNewCurrentFeeds calculates new supported feeds from current signal-total-powers.
func (k Keeper) CalculateNewCurrentFeeds(ctx sdk.Context) []types.Feed {
	signalTotalPowers := k.GetSignalTotalPowersByPower(ctx, k.GetParams(ctx).MaxCurrentFeeds)
	feeds := make([]types.Feed, 0, len(signalTotalPowers))
	params := k.GetParams(ctx)
	for _, signalTotalPower := range signalTotalPowers {
		interval := types.CalculateInterval(
			signalTotalPower.Power,
			params.PowerStepThreshold,
			params.MinInterval,
			params.MaxInterval,
		)
		if interval > 0 {
			feeds = append(
				feeds,
				types.NewFeed(
					signalTotalPower.ID,
					signalTotalPower.Power,
					interval,
				),
			)
		}
	}

	return feeds
}
