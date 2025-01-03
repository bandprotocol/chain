package keeper

import (
	"fmt"
	"sort"

	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/bandrng"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// GetMemberByAddress function retrieves a member of a group from the store by using address.
func (k Keeper) GetMemberByAddress(
	ctx sdk.Context,
	groupID tss.GroupID,
	address string,
) (types.Member, error) {
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		return types.Member{}, err
	}

	for _, member := range members {
		if member.Address == address {
			return member, nil
		}
	}

	return types.Member{}, types.ErrMemberNotFound.Wrapf(
		"failed to get member address %s from groupID %d", address, groupID,
	)
}

// GetAvailableMembers retrieves all members in the given group that are active and have an existing DE.
func (k Keeper) GetAvailableMembers(ctx sdk.Context, groupID tss.GroupID) []types.Member {
	iterator := k.GetGroupMembersIterator(ctx, groupID)
	defer iterator.Close()

	var availableMembers []types.Member
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)

		if !member.IsActive {
			continue
		}

		acc := sdk.MustAccAddressFromBech32(member.Address)
		if !k.HasDE(ctx, acc) {
			continue
		}

		availableMembers = append(availableMembers, member)
	}

	return availableMembers
}

// GetRandomMembers select a random members from the given group for a signing process.
// It selects a number of 'group.Threshold' assigned members out of the available members from
// the given group using a deterministic random number generator (DRBG).
func (k Keeper) GetRandomMembers(
	ctx sdk.Context,
	groupID tss.GroupID,
	nonce []byte,
) ([]types.Member, error) {
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get available members
	members := k.GetAvailableMembers(ctx, groupID)
	members_size := uint64(len(members))
	if group.Threshold > members_size {
		return nil, types.ErrInsufficientSigners.Wrapf(
			"the number of required signers %d is greater than available members %d",
			group.Threshold,
			members_size,
		)
	}

	// Create a deterministic random number generator (DRBG) using the rolling seed, signingID, and chain ID.
	rng, err := bandrng.NewRng(
		k.rollingseedKeeper.GetRollingSeed(ctx),
		nonce,
		[]byte(ctx.ChainID()),
	)
	if err != nil {
		return nil, types.ErrBadDrbgInitialization.Wrapf("fail to get rng: %v", err)
	}

	var selected []types.Member
	memberIdx := make([]int, members_size)
	for i := 0; i < int(members_size); i++ {
		memberIdx[i] = i
	}

	for i := uint64(0); i < group.Threshold; i++ {
		randomNumber := rng.NextUint64() % (members_size - i)

		// Swap the selected member with the last member in the list
		memberId := memberIdx[randomNumber]
		memberIdx[randomNumber] = memberIdx[members_size-i-1]

		// Append the selected member to the selected list
		selected = append(selected, members[memberId])
	}

	// Sort selected members
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })

	return selected, nil
}

// ValidateMemberID checks if the address is the given memberID of the group.
func (k Keeper) ValidateMemberID(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
	address string,
) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	if member.Address != address {
		return types.ErrInvalidMember.Wrapf(
			"memberID %d doesn't match with address %s in groupID %d", memberID, address, groupID,
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

// MarkMemberMalicious change member status to malicious.
func (k Keeper) MarkMemberMalicious(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}
	if member.IsMalicious {
		return nil
	}

	// update member status
	member.IsMalicious = true
	k.SetMember(ctx, member)
	return nil
}

// =====================================
// Member activation
// =====================================

// ActivateMember sets a boolean flag represent activeness of the user to true.
func (k Keeper) ActivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, true)
}

// DeactivateMember sets a boolean flag represent activeness of the user to false.
func (k Keeper) DeactivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error {
	return k.SetMemberIsActive(ctx, groupID, address, false)
}

// SetMemberIsActive sets a boolean flag represent activeness of the user.
func (k Keeper) SetMemberIsActive(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress, status bool) error {
	m, err := k.GetMemberByAddress(ctx, groupID, address.String())
	if err != nil {
		return err
	}

	m.IsActive = status
	k.SetMember(ctx, m)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSetMemberIsActive,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", m.ID)),
		sdk.NewAttribute(types.AttributeKeyMemberStatus, fmt.Sprintf("%t", status)),
	))

	return nil
}

// =====================================
// Member store
// =====================================

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
	iterator := k.GetGroupMembersIterator(ctx, groupID)
	defer iterator.Close()

	var members []types.Member
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
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MemberStoreKeyPrefix)
	defer iterator.Close()

	var members []types.Member
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}

	return members
}

// MustGetMembers retrieves all members of a group from the store. Panics error if not exists.
func (k Keeper) MustGetMembers(ctx sdk.Context, groupID tss.GroupID) []types.Member {
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		panic(err)
	}
	return members
}
