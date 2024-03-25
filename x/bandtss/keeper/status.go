package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// SetActiveStatuses sets the member status to active
func (k Keeper) SetActiveStatuses(ctx sdk.Context, addresses []sdk.AccAddress) error {
	statuses := make([]types.Status, 0, len(addresses))
	updatedAddress := make([]sdk.AccAddress, 0, len(addresses))
	isActives := make([]bool, 0, len(addresses))

	for _, addr := range addresses {
		status := k.GetStatus(ctx, addr)
		if status.Status == types.MEMBER_STATUS_ACTIVE {
			continue
		}

		params := k.GetParams(ctx)
		var penaltyDuration time.Duration
		if status.Status == types.MEMBER_STATUS_INACTIVE {
			penaltyDuration = params.InactivePenaltyDuration
		} else if status.Status == types.MEMBER_STATUS_JAIL {
			penaltyDuration = params.JailPenaltyDuration
		}

		if status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
			return types.ErrTooSoonToActivate
		}

		statuses = append(statuses, status)
		updatedAddress = append(updatedAddress, addr)
		isActives = append(isActives, true)
	}

	for _, status := range statuses {
		status.Status = types.MEMBER_STATUS_ACTIVE
		status.Since = ctx.BlockTime()
		status.LastActive = status.Since
		k.SetStatus(ctx, status)
	}

	k.tssKeeper.UpdateExistingMembersActiveness(ctx, k.GetCurrentGroupID(ctx), updatedAddress, isActives)

	return nil
}

// SetLastActive sets last active of the member
func (k Keeper) SetLastActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)

	if status.Status != types.MEMBER_STATUS_ACTIVE {
		return types.ErrInvalidStatus
	}

	status.LastActive = ctx.BlockTime()
	k.SetStatus(ctx, status)

	return nil
}

// SetInactive sets the member status to inactive
func (k Keeper) SetInactiveStatuses(ctx sdk.Context, addresses []sdk.AccAddress) {
	statuses := make([]types.Status, 0, len(addresses))
	updatedAddress := make([]sdk.AccAddress, 0, len(addresses))
	isActives := make([]bool, 0, len(addresses))

	for _, addr := range addresses {
		status := k.GetStatus(ctx, addr)

		// cannot overwrite jail status; NOTE: this does not cause an error.
		if status.Status == types.MEMBER_STATUS_INACTIVE || status.Status == types.MEMBER_STATUS_JAIL {
			continue
		}

		statuses = append(statuses, status)
		updatedAddress = append(updatedAddress, addr)
		isActives = append(isActives, false)
	}

	for _, status := range statuses {
		status.Status = types.MEMBER_STATUS_INACTIVE
		status.Since = ctx.BlockTime()
		k.SetStatus(ctx, status)
	}

	k.tssKeeper.UpdateExistingMembersActiveness(ctx, k.GetCurrentGroupID(ctx), updatedAddress, isActives)

	for _, addr := range updatedAddress {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeInactiveStatus,
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		))
	}
}

// SetJailStatuses sets the member status to jail
func (k Keeper) SetJailStatuses(ctx sdk.Context, addresses []sdk.AccAddress) {
	statuses := make([]types.Status, 0, len(addresses))
	updatedAddress := make([]sdk.AccAddress, 0, len(addresses))
	isActives := make([]bool, 0, len(addresses))

	for _, addr := range addresses {
		status := k.GetStatus(ctx, addr)

		// cannot overwrite jail status; NOTE: this does not cause an error.
		if status.Status == types.MEMBER_STATUS_JAIL {
			continue
		}

		statuses = append(statuses, status)
		updatedAddress = append(updatedAddress, addr)
		isActives = append(isActives, false)
	}

	for _, status := range statuses {
		status.Status = types.MEMBER_STATUS_JAIL
		status.Since = ctx.BlockTime()
		k.SetStatus(ctx, status)
	}

	k.tssKeeper.UpdateExistingMembersActiveness(ctx, k.GetCurrentGroupID(ctx), updatedAddress, isActives)

	for _, addr := range updatedAddress {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeJailStatus,
			sdk.NewAttribute(types.AttributeKeyAddress, addr.String()),
		))
	}
}
