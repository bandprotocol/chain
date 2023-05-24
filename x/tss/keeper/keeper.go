package keeper

import (
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
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
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
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetGroupCount function returns the current number of all groups ever existed.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey))
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
	groupID := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(tss.GroupID(groupID)), k.cdc.MustMarshal(&group))
	return groupID
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

// GetMember function retrieves a member of a group from the store.
func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, sdkerrors.Wrapf(
			types.ErrMemberNotFound,
			"failed to get member with groupID: %d and memberID: %d",
			groupID,
			memberID,
		)
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
func (k Keeper) VerifyMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, memberAddress string) bool {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil || member.Member != memberAddress {
		return false
	}
	return true
}

// SetRound1Data function sets round1 data for a member of a group.
func (k Keeper) SetRound1Data(ctx sdk.Context, groupID tss.GroupID, round1Data types.Round1Data) {
	// Add count
	k.AddRound1DataCount(ctx, groupID)
	ctx.KVStore(k.storeKey).
		Set(types.Round1DataMemberStoreKey(groupID, round1Data.MemberID), k.cdc.MustMarshal(&round1Data))
}

// GetRound1Data function retrieves round1 data of a member from the store.
func (k Keeper) GetRound1Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round1Data, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1DataMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Data{}, sdkerrors.Wrapf(
			types.ErrRound1DataNotFound,
			"failed to get round1 data with groupID: %d and memberID %d",
			groupID,
			memberID,
		)
	}
	var r1 types.Round1Data
	k.cdc.MustUnmarshal(bz, &r1)
	return r1, nil
}

// DeleteRound1Data removes the round1 data of a group member from the store.
func (k Keeper) DeleteRound1Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1DataMemberStoreKey(groupID, memberID))
}

// SetRound1DataCount sets the count of round1 data for a group in the store.
func (k Keeper) SetRound1DataCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round1DataCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound1DataCount retrieves the count of round1 data for a group from the store.
func (k Keeper) GetRound1DataCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1DataCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound1DataCount increments the count of round1 data for a group in the store.
func (k Keeper) AddRound1DataCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound1DataCount(ctx, groupID)
	k.SetRound1DataCount(ctx, groupID, count+1)
}

// GetRound1DataIterator function gets an iterator over all round1 data of a group.
func (k Keeper) GetRound1DataIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round1DataStoreKey(groupID))
}

// GetAllRound1Data retrieves all round1 data for a group from the store.
func (k Keeper) GetAllRound1Data(ctx sdk.Context, groupID tss.GroupID, groupSize uint64) []types.Round1Data {
	var allRound1Data []types.Round1Data
	iterator := k.GetRound1DataIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var round1Data types.Round1Data
		k.cdc.MustUnmarshal(iterator.Value(), &round1Data)
		allRound1Data = append(allRound1Data, round1Data)
	}
	return allRound1Data
}

// SetRound2Data method sets the round2Data of a member in the store and increments the count of round2Data.
func (k Keeper) SetRound2Data(
	ctx sdk.Context,
	groupID tss.GroupID,
	round2Data types.Round2Data,
) {
	// Add count
	k.AddRound2DataCount(ctx, groupID)

	ctx.KVStore(k.storeKey).
		Set(types.Round2DataMemberStoreKey(groupID, round2Data.MemberID), k.cdc.MustMarshal(&round2Data))
}

// GetRound2Data method retrieves the round2Data of a member from the store.
func (k Keeper) GetRound2Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round2Data, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2DataMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round2Data{}, sdkerrors.Wrapf(
			types.ErrRoundExpired,
			"failed to get round2Data with groupID: %d, memberID: %d",
			groupID,
			memberID,
		)
	}
	var r2 types.Round2Data
	k.cdc.MustUnmarshal(bz, &r2)
	return r2, nil
}

// DeleteRound2Data method deletes the round2Data of a member from the store.
func (k Keeper) DeleteRound2Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round2DataMemberStoreKey(groupID, memberID))
}

// SetRound2DataCount method sets the count of round2Datas in the store.
func (k Keeper) SetRound2DataCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round2DataCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound2DataCount method retrieves the count of round2Datas from the store.
func (k Keeper) GetRound2DataCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2DataCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound2DataCount method increments the count of round2Datas in the store.
func (k Keeper) AddRound2DataCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound2DataCount(ctx, groupID)
	k.SetRound2DataCount(ctx, groupID, count+1)
}

// GetRound2DataIterator function gets an iterator over all round1 data of a group.
func (k Keeper) GetRound2DataIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round2DataStoreKey(groupID))
}

// GetAllRound2Data method retrieves all round2Data for a given group from the store.
func (k Keeper) GetAllRound2Data(ctx sdk.Context, groupID tss.GroupID, groupSize uint64) []types.Round2Data {
	var allRound2Data []types.Round2Data
	iterator := k.GetRound2DataIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var round2Data types.Round2Data
		k.cdc.MustUnmarshal(iterator.Value(), &round2Data)
		allRound2Data = append(allRound2Data, round2Data)
	}
	return allRound2Data
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
