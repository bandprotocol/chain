package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// ActivateMember activates the member. This function returns an error if the given member is too
// soon to activate or the member is not in the current group.
func (k Keeper) ActivateMember(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)
	if status.Status == types.MEMBER_STATUS_ACTIVE {
		return nil
	}

	params := k.GetParams(ctx)
	if status.Status == types.MEMBER_STATUS_INACTIVE &&
		status.Since.Add(params.InactivePenaltyDuration).After(ctx.BlockTime()) {
		return types.ErrTooSoonToActivate
	}

	groupID := k.GetCurrentGroupID(ctx)
	if _, err := k.tssKeeper.GetMemberByAddress(ctx, groupID, address.String()); err != nil {
		return tsstypes.ErrMemberNotFound.Wrapf(
			"failed to get member with groupID: %d and address: %s",
			groupID,
			address,
		)
	}

	k.SetActiveStatus(ctx, address)
	return nil
}

// SetActiveStatus sets the member status to active. This function will panic if the given address
// isn't the member of the current group.
func (k Keeper) SetActiveStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := types.Status{
		Status:     types.MEMBER_STATUS_ACTIVE,
		Address:    address.String(),
		Since:      ctx.BlockTime(),
		LastActive: ctx.BlockTime(),
	}
	k.SetStatus(ctx, status)
	k.tssKeeper.MustSetMemberIsActive(ctx, k.GetCurrentGroupID(ctx), address, true)
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

// SetInactiveStatus sets the member status to inactive.  This function will panic if the given address
// isn't the member of the current group.
func (k Keeper) SetInactiveStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	// cannot overwrite jail status; NOTE: this does not cause an error.
	if status.Status == types.MEMBER_STATUS_INACTIVE {
		return
	}

	status.Status = types.MEMBER_STATUS_INACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, status)

	k.tssKeeper.MustSetMemberIsActive(ctx, k.GetCurrentGroupID(ctx), address, false)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInactiveStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))
}
