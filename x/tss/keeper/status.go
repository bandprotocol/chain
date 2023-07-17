package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetActive sets the member status to active
func (k Keeper) SetActive(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	if member.Status.IsActive {
		return nil
	}

	penaltyDuration := k.InactivePenaltyDuration(ctx)
	if member.Status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
		return types.ErrTooSoonToActivate
	}

	member.Status = types.MemberStatus{
		IsActive: true,
		Since:    ctx.BlockTime(),
	}
	k.SetMember(ctx, groupID, member)

	return nil
}

// SetInActive sets the member status to inactive
func (k Keeper) SetInActive(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	if !member.Status.IsActive {
		return nil
	}

	member.Status = types.MemberStatus{
		IsActive: false,
		Since:    ctx.BlockTime(),
	}

	k.SetMember(ctx, groupID, member)

	return nil
}
