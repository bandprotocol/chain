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
	sumDelegation := k.stakingKeeper.GetDelegatorBonded(ctx, delegator).Int64()
	if sumPower > sumDelegation {
		return types.ErrNotEnoughDelegation
	}
	return nil
}

// RemoveDelegatorSignals deletes signals and decrease feeds power of the signals of a delegator.
func (k Keeper) CalculateDelegatorSignalsPowerDiff(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
) map[string]int64 {
	signalIDToPowerDiff := make(map[string]int64)
	prevSignals := k.GetDelegatorSignals(ctx, delegator)
	k.DeleteDelegatorSignals(ctx, delegator)

	for _, signal := range prevSignals {
		signalIDToPowerDiff[signal.ID] -= signal.Power
	}

	// emit events for the removing signals operation.
	for _, signal := range prevSignals {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRemoveSignal,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
				sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
				sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	}

	k.SetDelegatorSignals(ctx, types.DelegatorSignals{Delegator: delegator.String(), Signals: signals})
	for _, signal := range signals {
		signalIDToPowerDiff[signal.ID] += signal.Power
	}
	for _, signal := range signals {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSubmitSignal,
				sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
				sdk.NewAttribute(types.AttributeKeySignalID, signal.ID),
				sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", signal.Power)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	}
	return signalIDToPowerDiff
}
