package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetFeedsIterator returns an iterator for feeds store.
func (k Keeper) GetFeedsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.FeedStoreKeyPrefix)
}

// GetFeeds returns a list of all feeds.
func (k Keeper) GetFeeds(ctx sdk.Context) (feeds []types.Feed) {
	iterator := k.GetFeedsIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var feed types.Feed
		k.cdc.MustUnmarshal(iterator.Value(), &feed)
		feeds = append(feeds, feed)
	}

	return
}

// GetFeed returns a feed by signal id.
func (k Keeper) GetFeed(ctx sdk.Context, signalID string) (types.Feed, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.FeedStoreKey(signalID))
	if bz == nil {
		return types.Feed{}, types.ErrFeedNotFound.Wrapf("failed to get feed detail for signal id: %s", signalID)
	}

	var f types.Feed
	k.cdc.MustUnmarshal(bz, &f)

	return f, nil
}

// SetFeeds sets multiple feeds.
func (k Keeper) SetFeeds(ctx sdk.Context, feeds []types.Feed) {
	for _, feed := range feeds {
		k.SetFeed(ctx, feed)
	}
}

// SetFeed sets a new feed to the feeds store or replace if feed with the same signal id existed.
func (k Keeper) SetFeed(ctx sdk.Context, feed types.Feed) {
	// set new timestamp if interval is updated
	prevFeed, err := k.GetFeed(ctx, feed.SignalID)
	k.deleteFeedByPowerIndex(ctx, prevFeed)
	if err == nil {
		if prevFeed.Interval != feed.Interval {
			feed.LastIntervalUpdateTimestamp = ctx.BlockTime().Unix()
		}
	}

	if feed.Power > 0 {
		ctx.KVStore(k.storeKey).Set(types.FeedStoreKey(feed.SignalID), k.cdc.MustMarshal(&feed))
		k.setFeedByPowerIndex(ctx, feed)
		emitEventUpdateFeed(ctx, feed)
	} else {
		k.DeleteFeed(ctx, feed)
		emitEventDeleteFeed(ctx, feed)
	}
}

// DeleteFeed deletes a feed from the feeds store.
func (k Keeper) DeleteFeed(ctx sdk.Context, feed types.Feed) {
	k.DeletePrice(ctx, feed.SignalID)
	k.deleteFeedByPowerIndex(ctx, feed)
	ctx.KVStore(k.storeKey).Delete(types.FeedStoreKey(feed.SignalID))
}

// setFeedByPowerIndex sets a feed in feedx by power index store.
func (k Keeper) setFeedByPowerIndex(ctx sdk.Context, feed types.Feed) {
	ctx.KVStore(k.storeKey).
		Set(types.FeedsByPowerIndexKey(feed.SignalID, feed.Power), []byte(feed.SignalID))
}

// DeleteFeedByPowerIndex deletes a feed from feedx by power index store.
func (k Keeper) deleteFeedByPowerIndex(ctx sdk.Context, feed types.Feed) {
	ctx.KVStore(k.storeKey).Delete(types.FeedsByPowerIndexKey(feed.SignalID, feed.Power))
}

// GetSupportedFeedsByPower gets the current group of bonded validators sorted by power-rank.
func (k Keeper) GetSupportedFeedsByPower(ctx sdk.Context) []types.Feed {
	maxFeeds := k.GetParams(ctx).MaxSupportedFeeds
	feeds := make([]types.Feed, maxFeeds)

	iterator := k.FeedsPowerStoreIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	i := 0
	for ; iterator.Valid() && i < int(maxFeeds); iterator.Next() {
		bz := iterator.Value()
		signalID := string(bz)
		feed, err := k.GetFeed(ctx, signalID)
		if err != nil || feed.Interval == 0 {
			continue
		}

		feeds[i] = feed
		i++
	}

	return feeds[:i] // trim
}

// FeedsPowerStoreIterator returns an iterator for feeds by power index store.
func (k Keeper) FeedsPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStoreReversePrefixIterator(ctx.KVStore(k.storeKey), types.FeedsByPowerIndexKeyPrefix)
}
