package keeper

import (
	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// SetMember sets a member of a group in the store.
func (k Keeper) SetMember(ctx sdk.Context, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberStoreKey(member.GroupID, member.ID), k.cdc.MustMarshal(&member))
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
		if member.IsAddress(address) {
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
	bz := ctx.KVStore(k.storeKey).Get(types.MemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, types.ErrMemberNotFound.Wrapf(
			"failed to get member with groupID: %d and memberID: %d",
			groupID,
			memberID,
		)
	}

	var member types.Member
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
func (k Keeper) GetGroupMembersIterator(ctx sdk.Context, groupID tss.GroupID) dbm.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
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
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MemberStoreKeyPrefix)
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
	ctx.KVStore(k.storeKey).Delete(types.MemberStoreKey(member.GroupID, member.ID))
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
	var availableMembers []types.Member
	iterator := k.GetGroupMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		if !member.IsActive {
			continue
		}

		acc, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
		}

		if !k.HasDE(ctx, acc) {
			continue
		}

		availableMembers = append(availableMembers, member)
	}

	if len(availableMembers) == 0 {
		return nil, types.ErrNoActiveMember.Wrapf("no active member in groupID: %d", groupID)
	}
	return availableMembers, nil
}

// SetMemberIsActive sets a boolean flag represent activeness of the user.
func (k Keeper) SetMemberIsActive(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress, status bool) error {
	m, err := k.GetMemberByAddress(ctx, groupID, address.String())
	if err != nil {
		return err
	}

	m.IsActive = status
	k.SetMember(ctx, m)
	return nil
}

// ActivateMember sets a boolean flag represent activeness of the user to true.
func (k Keeper) ActivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, true)
}

// DeactivateMember sets a boolean flag represent activeness of the user to false.
func (k Keeper) DeactivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, false)
}

// ValidateMemberID checks if the address is the given memberID of the group.
func (k Keeper) ValidateMemberID(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
	address string,
) error {
	// Get member and verify if the sender is in the group
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	if !member.IsAddress(address) {
		return types.ErrMemberNotAuthorized.Wrapf(
			"memberID %d address %s is not match in this group",
			memberID,
			address,
		)
	}

	return nil
}

// UpdateMemberPubKey computes own public key and set it to the member.
func (k Keeper) UpdateMemberPubKey(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	// Compute own public key
	accCommits := k.GetAllAccumulatedCommits(ctx, groupID)
	ownPubKey, err := tss.ComputeOwnPublicKey(accCommits, memberID)
	if err != nil {
		return types.ErrComputeOwnPubKeyFailed.Wrapf("compute own public key failed; %s", err)
	}

	// set own public key to member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	member.PubKey = ownPubKey
	k.SetMember(ctx, member)
	return nil
}
