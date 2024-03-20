package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (k Keeper) GetFeedsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.FeedStoreKeyPrefix)
}

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

	return feeds
}

func (k Keeper) GetFeed(ctx sdk.Context, signalID string) (types.Feed, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.FeedStoreKey(signalID))
	if bz == nil {
		return types.Feed{}, types.ErrFeedNotFound.Wrapf("failed to get feed detail for signal id: %s", signalID)
	}

	var f types.Feed
	k.cdc.MustUnmarshal(bz, &f)

	return f, nil
}

func (k Keeper) SetFeeds(ctx sdk.Context, feeds []types.Feed) {
	for _, feed := range feeds {
		k.SetFeed(ctx, feed)
	}
}

func (k Keeper) SetFeed(ctx sdk.Context, feed types.Feed) {
	ctx.KVStore(k.storeKey).Set(types.FeedStoreKey(feed.SignalID), k.cdc.MustMarshal(&feed))
}

func (k Keeper) DeleteFeed(ctx sdk.Context, signalID string) {
	k.DeletePrice(ctx, signalID)
	ctx.KVStore(k.storeKey).Delete(types.FeedStoreKey(signalID))
}

func (k Keeper) SetFeedsByPowerIndex(ctx sdk.Context, feeds []types.Feed) {
	for _, feed := range feeds {
		k.SetFeedByPowerIndex(ctx, feed)
	}
}

func (k Keeper) SetFeedByPowerIndex(ctx sdk.Context, feed types.Feed) {
	ctx.KVStore(k.storeKey).
		Set(types.FeedsByPowerIndexKey(feed.SignalID, feed.Power), []byte(feed.SignalID))
}

func (k Keeper) DeleteFeedByPowerIndex(ctx sdk.Context, feed types.Feed) {
	ctx.KVStore(k.storeKey).Delete(types.FeedsByPowerIndexKey(feed.SignalID, feed.Power))
}

// GetSupportedFeedsByPower gets the current group of bonded validators sorted by power-rank
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

func (k Keeper) FeedsPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStoreReversePrefixIterator(ctx.KVStore(k.storeKey), types.FeedsByPowerIndexKeyPrefix)
}
