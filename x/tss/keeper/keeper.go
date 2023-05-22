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

// SetGroupCount sets the number of group count to the given value.
func (k Keeper) SetGroupCount(ctx sdk.Context, count uint64) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, bz)
}

// GetGroupCount returns the current number of all groups ever exist.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey)
	return binary.BigEndian.Uint64(bz)
}

// GetNextGroupID increments and returns the current number of groups.
func (k Keeper) GetNextGroupID(ctx sdk.Context) tss.GroupID {
	groupNumber := k.GetGroupCount(ctx)
	k.SetGroupCount(ctx, groupNumber+1)
	return tss.GroupID(groupNumber + 1)
}

// IsGrantee checks if the granter granted to the grantee
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

func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) tss.GroupID {
	id := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(tss.GroupID(id)), k.cdc.MustMarshal(&group))
	return id
}

func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get group with groupID: %d", groupID)
	}

	group := types.Group{}
	k.cdc.MustUnmarshal(bz, &group)
	return group, nil
}

func (k Keeper) UpdateGroup(ctx sdk.Context, groupID tss.GroupID, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(groupID), k.cdc.MustMarshal(&group))
}

func (k Keeper) SetDKGContext(ctx sdk.Context, groupID tss.GroupID, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

func (k Keeper) GetDKGContext(ctx sdk.Context, groupID tss.GroupID) ([]byte, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDKGContextNotFound, "failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

func (k Keeper) SetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

func (k Keeper) SetMembers(ctx sdk.Context, groupID tss.GroupID, members []types.Member) {
	for i, m := range members {
		ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, tss.MemberID(i+1)), k.cdc.MustMarshal(&m))
	}
}

func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, sdkerrors.Wrapf(types.ErrMemberNotFound, "failed to get member with groupID: %d and memberID: %d", groupID, memberID)
	}

	member := types.Member{}
	k.cdc.MustUnmarshal(bz, &member)
	return member, nil
}

func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

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

func (k Keeper) SetRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, round1Commitment types.Round1Commitment) {
	ctx.KVStore(k.storeKey).Set(types.Round1CommitmentMemberStoreKey(groupID, memberID), k.cdc.MustMarshal(&round1Commitment))
}

func (k Keeper) GetRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round1Commitment, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1CommitmentMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Commitment{}, sdkerrors.Wrapf(types.ErrRound1CommitmentsNotFound, "failed to get round 1 commitments with groupID: %d and memberID %d", groupID, memberID)
	}
	var r1c types.Round1Commitment
	k.cdc.MustUnmarshal(bz, &r1c)
	return r1c, nil
}

func (k Keeper) DeleteRound1Commitment(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1CommitmentMemberStoreKey(groupID, memberID))
}

func (k Keeper) getRound1CommitmentsIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round1CommitmentStoreKey(groupID))
}

func (k Keeper) GetRound1CommitmentsCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	var count uint64
	iterator := k.getRound1CommitmentsIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count += 1
	}
	return count
}

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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
