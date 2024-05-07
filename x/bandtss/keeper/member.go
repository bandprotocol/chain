package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// ActivateMember activates the member. This function returns an error if the given member is too
// soon to activate or the member is not in the current group.
func (k Keeper) ActivateMember(ctx sdk.Context, address sdk.AccAddress) error {
	member, err := k.GetMember(ctx, address)
	if err != nil {
		return err
	}

	if member.IsActive {
		return types.ErrMemberAlreadyActive
	}

	if member.Since.Add(k.GetParams(ctx).InactivePenaltyDuration).After(ctx.BlockTime()) {
		return types.ErrTooSoonToActivate
	}

	member.IsActive = true
	member.LastActive = ctx.BlockTime()
	k.SetMember(ctx, member)

	groupID := k.GetCurrentGroupID(ctx)
	return k.tssKeeper.ActivateMember(ctx, groupID, address)
}

// AddNewMember adds a new member to the group and return error if already exists
func (k Keeper) AddNewMember(ctx sdk.Context, address sdk.AccAddress) error {
	if k.HasMember(ctx, address) {
		return types.ErrMemberAlreadyExists.Wrapf("address : %v", address)
	}

	member := types.Member{
		Address:    address.String(),
		IsActive:   true,
		Since:      ctx.BlockTime(),
		LastActive: ctx.BlockTime(),
	}
	k.SetMember(ctx, member)

	return nil
}

// SetLastActive sets last active of the member
func (k Keeper) SetLastActive(ctx sdk.Context, address sdk.AccAddress) error {
	member, err := k.GetMember(ctx, address)
	if err != nil {
		return err
	}

	if !member.IsActive {
		return types.ErrInvalidStatus
	}

	member.LastActive = ctx.BlockTime()
	k.SetMember(ctx, member)

	return nil
}

// DeactivateMember flags is_active to false. This function will panic if the given address
// isn't the member of the current group.
func (k Keeper) DeactivateMember(ctx sdk.Context, address sdk.AccAddress) error {
	member, err := k.GetMember(ctx, address)
	if err != nil {
		return err
	}

	if !member.IsActive {
		return nil
	}

	member.IsActive = false
	member.Since = ctx.BlockTime()
	k.SetMember(ctx, member)

	if err := k.tssKeeper.DeactivateMember(ctx, k.GetCurrentGroupID(ctx), address); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeInactiveStatus,
		sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
	))

	return nil
}
