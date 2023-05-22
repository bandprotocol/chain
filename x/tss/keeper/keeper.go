package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	authzKeeper types.AuthzKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,

	authzKeeper types.AuthzKeeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		authzKeeper: authzKeeper,
	}
}

// SetGroupCount function sets the number of group count to the given value.
func (k Keeper) SetGroupCount(ctx sdk.Context, count uint64) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, bz)
}

// GetGroupCount function returns the current number of all groups ever existed.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey)
	return binary.BigEndian.Uint64(bz)
}

// GetNextGroupID function increments the group count and returns the current number of groups.
func (k Keeper) GetNextGroupID(ctx sdk.Context) tss.GroupID {
	groupNumber := k.GetGroupCount(ctx)
	k.SetGroupCount(ctx, groupNumber+1)
	return tss.GroupID(groupNumber + 1)
}

// IsGrantee function checks if the granter granted permissions to the grantee.
func (k Keeper) IsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
	for _, msg := range types.MsgGrants {
		cap, _ := k.authzKeeper.GetAuthorization(
			ctx,
			grantee,
			granter,
			msg,
		)

		if cap == nil {
			return false
		}
	}

	return true
}

// CreateNewGroup function creates a new group in the store and returns the id of the group.
func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) tss.GroupID {
	id := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(tss.GroupID(id)), k.cdc.MustMarshal(&group))
	return id
}

// GetGroup function retrieves a group from the store.
func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get group with groupID: %d", groupID)
	}

	group := types.Group{}
	k.cdc.MustUnmarshal(bz, &group)
	return group, nil
}

// UpdateGroup function updates a group in the store.
func (k Keeper) UpdateGroup(ctx sdk.Context, groupID tss.GroupID, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(groupID), k.cdc.MustMarshal(&group))
}

// SetDKGContext function sets DKG context for a group in the store.
func (k Keeper) SetDKGContext(ctx sdk.Context, groupID tss.GroupID, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

// GetDKGContext function retrieves DKG context of a group from the store.
func (k Keeper) GetDKGContext(ctx sdk.Context, groupID tss.GroupID) ([]byte, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDKGContextNotFound, "failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

// SetMember function sets a member of a group in the store.
func (k Keeper) SetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

// SetMembers function sets members of a group in the store.
func (k Keeper) SetMembers(ctx sdk.Context, groupID tss.GroupID, members []types.Member) {
	for i, m := range members {
		ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, tss.MemberID(i+1)), k.cdc.MustMarshal(&m))
	}
}

// GetMember function retrieves a member of a group from the store.
func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, sdkerrors.Wrapf(types.ErrMemberNotFound, "failed to get member with groupID: %d and memberID: %d", groupID, memberID)
	}

	member := types.Member{}
	k.cdc.MustUnmarshal(bz, &member)
	return member, nil
}

// GetMembersIterator function gets an iterator over all members of a group.
func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

// GetMembers function retrieves all members of a group from the store.
func (k Keeper) GetMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var members []types.Member
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}
	if len(members) == 0 {
		return nil, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// VerifyMember function verifies if a member is part of a group.
func (k Keeper) VerifyMember(ctx sdk.Context, groupID tss.GroupID, memberAddress string) (tss.MemberID, error) {
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return 0, err
	}

	for i, m := range members {
		if m.Signer == memberAddress {
			return tss.MemberID(i + 1), nil
		}
	}
	return 0, sdkerrors.Wrapf(types.ErrMemberNotAuthorized, "failed to get member %s on groupID %d", memberAddress, groupID)
}

// SetRound1Commitment function sets round 1 commitment for a member of a group.
func (k Keeper) SetRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, round1Commitment types.Round1Commitment) {
	// Add count
	k.AddRound1CommitmentsCount(ctx, groupID)

	ctx.KVStore(k.storeKey).Set(types.Round1CommitmentMemberStoreKey(groupID, memberID), k.cdc.MustMarshal(&round1Commitment))
}

// GetRound1Commitment function retrieves round 1 commitment of a member from the store.
func (k Keeper) GetRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round1Commitment, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1CommitmentMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Commitment{}, sdkerrors.Wrapf(types.ErrRound1CommitmentsNotFound, "failed to get round 1 commitments with groupID: %d and memberID %d", groupID, memberID)
	}
	var r1c types.Round1Commitment
	k.cdc.MustUnmarshal(bz, &r1c)
	return r1c, nil
}

// DeleteRound1Commitment removes the round 1 commitment of a group member from the store.
func (k Keeper) DeleteRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1CommitmentMemberStoreKey(groupID, memberID))
}

// SetRound1CommitmentsCount sets the count of round 1 commitments for a group in the store.
func (k Keeper) SetRound1CommitmentsCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round1CommitmentsCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound1CommitmentsCount retrieves the count of round 1 commitments for a group from the store.
func (k Keeper) GetRound1CommitmentsCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1CommitmentsCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound1CommitmentsCount increments the count of round 1 commitments for a group in the store.
func (k Keeper) AddRound1CommitmentsCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound1CommitmentsCount(ctx, groupID)
	k.SetRound1CommitmentsCount(ctx, groupID, count+1)
}

// GetAllRound1Commitments retrieves all round 1 commitments for a group from the store.
func (k Keeper) GetAllRound1Commitments(ctx sdk.Context, groupID tss.GroupID, groupSize uint64) []*types.Round1Commitment {
	allRound1Commitments := make([]*types.Round1Commitment, groupSize)
	for i := uint64(1); i <= groupSize; i++ {
		round1Commitment, err := k.GetRound1Commitment(ctx, groupID, tss.MemberID(i))
		if err != nil {
			// allRound1Commitments array start at 0
			allRound1Commitments[i-1] = nil
		} else {
			// allRound1Commitments array start at 0
			allRound1Commitments[i-1] = &round1Commitment
		}
	}

	return allRound1Commitments
}

// SetRound2Share method sets the round 2 share of a member in the store and increments the count of round 2 shares.
func (k Keeper) SetRound2Share(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, round2Share types.Round2Share) {
	// Add count
	k.AddRound2SharesCount(ctx, groupID)

	ctx.KVStore(k.storeKey).Set(types.Round2ShareMemberStoreKey(groupID, memberID), k.cdc.MustMarshal(&round2Share))
}

// GetRound2Share method retrieves the round 2 share of a member from the store.
func (k Keeper) GetRound2Share(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round2Share, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2ShareMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round2Share{}, sdkerrors.Wrapf(types.ErrRound2ShareNotFound, "failed to get round 2 share with groupID: %d, memberID: %d", groupID, memberID)
	}
	var r2s types.Round2Share
	k.cdc.MustUnmarshal(bz, &r2s)
	return r2s, nil
}

// DeleteRound2share method deletes the round 2 share of a member from the store.
func (k Keeper) DeleteRound2share(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round2ShareMemberStoreKey(groupID, memberID))
}

// SetRound2SharesCount method sets the count of round 2 shares in the store.
func (k Keeper) SetRound2SharesCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round2ShareCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound2SharesCount method retrieves the count of round 2 shares from the store.
func (k Keeper) GetRound2SharesCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2ShareCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound2SharesCount method increments the count of round 2 shares in the store.
func (k Keeper) AddRound2SharesCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound2SharesCount(ctx, groupID)
	k.SetRound2SharesCount(ctx, groupID, count+1)
}

// GetAllRound2Shares method retrieves all round 2 shares for a given group from the store.
func (k Keeper) GetAllRound2Shares(ctx sdk.Context, groupID tss.GroupID, groupSize uint64) []*types.Round2Share {
	allRound2Shares := make([]*types.Round2Share, groupSize)
	for i := uint64(1); i <= groupSize; i++ {
		round2Share, err := k.GetRound2Share(ctx, groupID, tss.MemberID(i))
		if err != nil {
			// allRound2Shares array start at 0
			allRound2Shares[i-1] = nil
		} else {
			// allRound2Shares array start at 0
			allRound2Shares[i-1] = &round2Share
		}
	}
	return allRound2Shares
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
