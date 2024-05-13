package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetMember sets a member of a group in the store.
func (k Keeper) SetMember(ctx sdk.Context, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(member.GroupID, member.ID), k.cdc.MustMarshal(&member))
}

// SetMembers sets members of a group in the store.
func (k Keeper) SetMembers(ctx sdk.Context, members []types.Member) {
	for _, member := range members {
		k.SetMember(ctx, member)
	}
}

// GetMemberByAddress function retrieves a member of a group from the store by using address.
func (k Keeper) GetMemberByAddress(ctx sdk.Context, groupID tss.GroupID, address string) (types.Member, error) {
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		return types.Member{}, err
	}

	for _, member := range members {
		if member.Verify(address) {
			return member, nil
		}
	}

	return types.Member{}, types.ErrMemberNotFound.Wrapf(
		"failed to get member with groupID: %d and address: %s",
		groupID,
		address,
	)
}

// GetMember function retrieves a member of a group from the store.
func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, types.ErrMemberNotFound.Wrapf(
			"failed to get member with groupID: %d and memberID: %d",
			groupID,
			memberID,
		)
	}

	member := types.Member{}
	k.cdc.MustUnmarshal(bz, &member)
	return member, nil
}

// MustGetMember returns the member for the given groupID and memberID. Panics error if not exists.
func (k Keeper) MustGetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) types.Member {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		panic(err)
	}
	return member
}

// GetGroupMembersIterator gets an iterator over all members of a group.
func (k Keeper) GetGroupMembersIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

// GetGroupMembers retrieves all members of a group from the store.
func (k Keeper) GetGroupMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var members []types.Member
	iterator := k.GetGroupMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}
	if len(members) == 0 {
		return nil, types.ErrMemberNotFound.Wrapf("failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// GetMembers retrieves all members from store.
func (k Keeper) GetMembers(ctx sdk.Context) []types.Member {
	var members []types.Member
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MemberStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}

	return members
}

// DeleteGroupMembers removes all members in the group
func (k Keeper) DeleteGroupMembers(ctx sdk.Context, groupID tss.GroupID) error {
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		return err
	}

	for _, member := range members {
		k.DeleteMember(ctx, member)
	}

	return nil
}

// DeleteMember removes a member
func (k Keeper) DeleteMember(ctx sdk.Context, member types.Member) {
	ctx.KVStore(k.storeKey).Delete(types.MemberOfGroupKey(member.GroupID, member.ID))
}

// MustGetMembers retrieves all members of a group from the store. Panics error if not exists.
func (k Keeper) MustGetMembers(ctx sdk.Context, groupID tss.GroupID) []types.Member {
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		panic(err)
	}
	return members
}

// GetAvailableMembers retrieves all active members of a group from the store.
func (k Keeper) GetAvailableMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var activeMembers []types.Member
	iterator := k.GetGroupMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)

		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}

	// Filter members that have DE left
	filteredMembers, err := k.FilterMembersHaveDE(ctx, activeMembers)
	if err != nil {
		return nil, err
	}

	if len(filteredMembers) == 0 {
		return nil, types.ErrNoActiveMember.Wrapf("no active member in groupID: %d", groupID)
	}
	return filteredMembers, nil
}

// SetMemberIsActive sets a boolean flag represent activeness of the user.
func (k Keeper) SetMemberIsActive(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress, status bool) error {
	members := k.MustGetMembers(ctx, groupID)
	for _, m := range members {
		if m.Address == address.String() {
			m.IsActive = status
			k.SetMember(ctx, m)
			return nil
		}
	}

	return types.ErrMemberNotFound.Wrapf(
		"failed to set member active status with groupID: %d and address: %s",
		groupID,
		address,
	)
}

// ActivateMember sets a boolean flag represent activeness of the user to true.
func (k Keeper) ActivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, true)
}

// DeactivateMember sets a boolean flag represent activeness of the user to false.
func (k Keeper) DeactivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, false)
}
