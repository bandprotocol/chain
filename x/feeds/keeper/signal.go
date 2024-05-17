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
	sumDelegation := k.stakingKeeper.GetDelegatorBonded(ctx, delegator).Int64()
	if sumPower > sumDelegation {
		return types.ErrNotEnoughDelegation
	}

	return nil
}

// CalculateDelegatorSignalsPowerDiff calculates feed power differences from delegator's previous signals and new signals.
func (k Keeper) CalculateDelegatorSignalsPowerDiff(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
) (map[string]int64, error) {
	signalIDToPowerDiff := make(map[string]int64)

	prevSignals := k.GetDelegatorSignals(ctx, delegator)
	k.DeleteDelegatorSignals(ctx, delegator)

	for _, prevSignal := range prevSignals {
		signalIDToPowerDiff[prevSignal.ID] -= prevSignal.Power
	}

	k.SetDelegatorSignals(ctx, types.DelegatorSignals{Delegator: delegator.String(), Signals: signals})

	for _, signal := range signals {
		if signal.ID == "" || signal.Power <= 0 {
			return nil, types.ErrInvalidSignal.Wrap(
				"signal id cannot be empty and its power must be positive",
			)
		}
		signalIDToPowerDiff[signal.ID] += signal.Power
	}

	return signalIDToPowerDiff, nil
}
