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

	left := k.tssKeeper.GetDECount(ctx, address)
	if left == 0 {
		status.Status = types.MEMBER_STATUS_PAUSED
	} else {
		status.Status = types.MEMBER_STATUS_ACTIVE
	}

	status.Address = address.String()
	status.Since = ctx.BlockTime()
	status.LastActive = status.Since
	k.SetMemberStatus(ctx, status)
	k.tssKeeper.SetMemberStatus(ctx, address, true)

	return nil
}

// SetLastActive sets last active of the member
func (k Keeper) SetLastActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)

	if status.Status != types.MEMBER_STATUS_ACTIVE && status.Status != types.MEMBER_STATUS_PAUSED {
		return types.ErrInvalidStatus
	}

	status.LastActive = ctx.BlockTime()
	k.SetMemberStatus(ctx, status)
	k.tssKeeper.SetMemberStatus(ctx, address, true)

	return nil
}

// SetInactive sets the member status to inactive
func (k Keeper) SetInactiveStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	if status.Status == types.MEMBER_STATUS_INACTIVE {
		return
	} else if status.Status == types.MEMBER_STATUS_JAIL {
		return
	}

	status.Status = types.MEMBER_STATUS_INACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetMemberStatus(ctx, status)
	k.tssKeeper.SetMemberStatus(ctx, address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInactiveStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))
}

// SetPaused sets the member status to paused
func (k Keeper) SetPausedStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	if status.Status != types.MEMBER_STATUS_PAUSED {
		return
	}

	status.Status = types.MEMBER_STATUS_PAUSED
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetMemberStatus(ctx, status)
	k.tssKeeper.SetMemberStatus(ctx, address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePausedStatus,
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
	k.SetMemberStatus(ctx, status)
	k.tssKeeper.SetMemberStatus(ctx, address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeJailStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))
}
