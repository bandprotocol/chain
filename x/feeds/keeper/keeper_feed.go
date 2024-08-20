package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
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
	cf := types.CurrentFeeds{
		Feeds:               feeds,
		LastUpdateTimestamp: ctx.BlockTime().Unix(),
		LastUpdateBlock:     ctx.BlockHeight(),
	}

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
			feed := types.Feed{
				SignalID: signalTotalPower.ID,
				Interval: interval,
				Power:    signalTotalPower.Power,
			}
			feeds = append(feeds, feed)
		}
	}

	return feeds
}
