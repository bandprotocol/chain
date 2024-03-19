package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// CheckDelegatorDelegation checks whether the delegator has enough delegation for signals.
func (k Keeper) CheckDelegatorDelegation(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
) error {
	sumPower := sumPower(signals)
	sumDelegation := k.stakingKeeper.GetDelegatorBonded(ctx, delegator).Uint64()
	if sumPower > sumDelegation {
		return types.ErrNotEnoughDelegation
	}
	return nil
}

// RemoveDelegatorSignal deletes previous signals from delegator and decrease feed power by the previous signals.
func (k Keeper) RemoveDelegatorPreviousSignals(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signalIDToIntervalDiff map[string]int64,
) (map[string]int64, error) {
	prevSignals := k.GetDelegatorSignals(ctx, delegator)
	for _, prevSignal := range prevSignals {
		feed, err := k.GetFeed(ctx, prevSignal.ID)
		if err != nil {
			return nil, err
		}
		// before changing in feed, delete the FeedByPower index
		k.DeleteFeedByPowerIndex(ctx, feed)

		feed.Power -= prevSignal.Power
		prevInterval := feed.Interval
		feed.Interval = calculateInterval(int64(feed.Power), k.GetParams(ctx))
		k.SetFeed(ctx, feed)

		// setting FeedByPowerIndex every time setting feed
		k.SetFeedByPowerIndex(ctx, feed)

		intervalDiff := (feed.Interval - prevInterval) + signalIDToIntervalDiff[feed.SignalID]
		if intervalDiff == 0 {
			delete(signalIDToIntervalDiff, feed.SignalID)
		} else {
			signalIDToIntervalDiff[feed.SignalID] = intervalDiff
		}
	}
	// return intervaldiff of signal ids
	return signalIDToIntervalDiff, nil
}

// RegisterDelegatorSignals increases feed power by the new signals.
func (k Keeper) RegisterDelegatorSignals(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
	signalIDToIntervalDiff map[string]int64,
) (map[string]int64, error) {
	k.SetDelegatorSignals(ctx, delegator, types.Signals{Signals: signals})
	for _, signal := range signals {
		feed, err := k.GetFeed(ctx, signal.ID)
		if err != nil {
			feed = types.Feed{
				SignalID:                    signal.ID,
				Power:                       0,
				Interval:                    0,
				LastIntervalUpdateTimestamp: 0,
			}
		}
		// before changing in feed, delete the FeedByPower index
		k.DeleteFeedByPowerIndex(ctx, feed)

		feed.Power += signal.Power
		prevInterval := feed.Interval
		feed.Interval = calculateInterval(int64(feed.Power), k.GetParams(ctx))
		k.SetFeed(ctx, feed)

		// setting FeedByPowerIndex every time setting feed
		k.SetFeedByPowerIndex(ctx, feed)

		// if the sum interval differences is zero then the interval is not changed
		intervalDiff := (feed.Interval - prevInterval) + signalIDToIntervalDiff[feed.SignalID]
		if intervalDiff == 0 {
			delete(signalIDToIntervalDiff, feed.SignalID)
		} else {
			signalIDToIntervalDiff[feed.SignalID] = intervalDiff
		}
	}
	return signalIDToIntervalDiff, nil
}

// UpdateFeedIntervalTimestamp updates the interval timestamp for feeds where the interval has changed.
func (k Keeper) UpdateFeedIntervalTimestamp(
	ctx sdk.Context,
	signalIDToIntervalDiff map[string]int64,
) error {
	for signalID := range signalIDToIntervalDiff {
		feed, err := k.GetFeed(ctx, signalID)
		if err != nil {
			return err
		}
		// before changing in feed, delete the FeedByPower index
		k.DeleteFeedByPowerIndex(ctx, feed)

		feed.LastIntervalUpdateTimestamp = ctx.BlockTime().Unix()
		k.SetFeed(ctx, feed)

		// setting FeedByPowerIndex every time setting feed
		k.SetFeedByPowerIndex(ctx, feed)
	}
	return nil
}
