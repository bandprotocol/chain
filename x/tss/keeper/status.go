package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// SetActive sets the member status to active
func (k Keeper) SetActive(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	member.IsActive = true
	k.SetMember(ctx, groupID, member)

	return nil
}

// SetInActive sets the member status to inactive
func (k Keeper) SetInActive(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	member.IsActive = false
	k.SetMember(ctx, groupID, member)

	return nil
}
