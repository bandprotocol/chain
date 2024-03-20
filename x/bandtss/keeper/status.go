package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// SetActive sets the member status to active
func (k Keeper) SetActiveStatus(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)
	if status.Status == types.MEMBER_STATUS_ACTIVE {
		return nil
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

	status.Status = types.MEMBER_STATUS_ACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	status.LastActive = status.Since
	k.SetStatus(ctx, status)
	k.tssKeeper.SetMemberIsActive(ctx, address, status.Status == types.MEMBER_STATUS_ACTIVE)

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
func (k Keeper) SetInactiveStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	// cannot overwrite jail status; NOTE: this does not cause an error.
	if status.Status == types.MEMBER_STATUS_INACTIVE || status.Status == types.MEMBER_STATUS_JAIL {
		return
	}

	status.Status = types.MEMBER_STATUS_INACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, status)
	k.tssKeeper.SetMemberIsActive(ctx, address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInactiveStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))
}

// SetJail sets the member status to jail
func (k Keeper) SetJailStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	if status.Status == types.MEMBER_STATUS_JAIL {
		return
	}

	status.Status = types.MEMBER_STATUS_JAIL
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, status)
	k.tssKeeper.SetMemberIsActive(ctx, address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeJailStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))
}
