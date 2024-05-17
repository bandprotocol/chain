package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetSupportedFeeds gets the current supported feeds.
func (k Keeper) GetSupportedFeeds(ctx sdk.Context) (sp types.SupportedFeeds) {
	bz := ctx.KVStore(k.storeKey).Get(types.SupportedFeedsStoreKey)
	if bz == nil {
		return sp
	}

	k.cdc.MustUnmarshal(bz, &sp)

	return sp
}

// SetSupportedFeeds sets new supported feeds to the store.
func (k Keeper) SetSupportedFeeds(ctx sdk.Context, feeds []types.Feed) {
	sf := types.SupportedFeeds{
		Feeds:               feeds,
		LastUpdateTimestamp: ctx.BlockTime().Unix(),
		LastUpdateBlock:     ctx.BlockHeight(),
	}

	ctx.KVStore(k.storeKey).Set(types.SupportedFeedsStoreKey, k.cdc.MustMarshal(&sf))
	emitEventUpdateSupportedFeeds(ctx, sf)
}

// CalculateNewSupportedFeeds calculates new supported feeds from current signal-total-powers.
func (k Keeper) CalculateNewSupportedFeeds(ctx sdk.Context) (feeds []types.Feed) {
	signalTotalPowers := k.GetSignalTotalPowersByPower(ctx, k.GetParams(ctx).MaxSupportedFeeds)
	for _, signalTotalPower := range signalTotalPowers {
		interval, deviationInThousandth := CalculateIntervalAndDeviation(
			signalTotalPower.Power,
			k.GetParams(ctx),
		)
		feed := types.Feed{
			SignalID:              signalTotalPower.ID,
			Interval:              interval,
			DeviationInThousandth: deviationInThousandth,
		}
		feeds = append(feeds, feed)
	}

	return
}
