package keeper

import (
	"fmt"

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

// RemoveSignal deletes previous signals from  and decrease feed power by the previous signals.
func (k Keeper) RemovePreviousSignals(
	ctx sdk.Context,
	signals []types.Signal,
	signalIDToIntervalDiff map[string]int64,
) (map[string]int64, error) {
	for _, signal := range signals {
		feed, err := k.GetFeed(ctx, signal.ID)
		if err != nil {
			return nil, err
		}
		// before changing in feed, delete the FeedByPower index
		k.DeleteFeedByPowerIndex(ctx, feed)

		feed.Power -= signal.Power
		prevInterval := feed.Interval
		feed.Interval, feed.DeviationInThousandth = calculateIntervalAndDeviation(int64(feed.Power), k.GetParams(ctx))
		k.SetFeed(ctx, feed)

		intervalDiff := (feed.Interval - prevInterval) + signalIDToIntervalDiff[feed.SignalID]
		if intervalDiff == 0 {
			delete(signalIDToIntervalDiff, feed.SignalID)
		} else {
			signalIDToIntervalDiff[feed.SignalID] = intervalDiff
		}
	}
	// emit events for the removing signals operation.
	for _, signal := range signals {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRemoveSignals,
				sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
				sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
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
	k.SetDelegatorSignals(ctx, types.DelegatorSignals{Delegator: delegator.String(), Signals: signals})
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
		feed.Interval, feed.DeviationInThousandth = calculateIntervalAndDeviation(int64(feed.Power), k.GetParams(ctx))
		k.SetFeed(ctx, feed)

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
) {
	for signalID := range signalIDToIntervalDiff {
		feed, err := k.GetFeed(ctx, signalID)
		if err != nil {
			// if feed is deleted, no need to update its timestamp
			continue
		}
		// before changing in feed, delete the FeedByPower index
		k.DeleteFeedByPowerIndex(ctx, feed)

		feed.LastIntervalUpdateTimestamp = ctx.BlockTime().Unix()
		k.SetFeed(ctx, feed)
	}
}
