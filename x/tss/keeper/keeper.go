package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

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
func (k Keeper) GetNextGroupID(ctx sdk.Context) uint64 {
	groupNumber := k.GetGroupCount(ctx)
	k.SetGroupCount(ctx, groupNumber+1)
	return groupNumber + 1
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

func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) uint64 {
	id := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(id), k.cdc.MustMarshal(&group))
	return id
}

func (k Keeper) GetGroup(ctx sdk.Context, groupID uint64) (types.Group, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, false
	}

	group := types.Group{}
	k.cdc.MustUnmarshal(bz, &group)
	return group, true
}

func (k Keeper) UpdateGroup(ctx sdk.Context, groupID uint64, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(groupID), k.cdc.MustMarshal(&group))
}

func (k Keeper) SetDKGContext(ctx sdk.Context, groupID uint64, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

func (k Keeper) GetDKGContext(ctx sdk.Context, groupID uint64) ([]byte, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, false
	}
	return bz, true
}

func (k Keeper) SetMember(ctx sdk.Context, groupID, memberID uint64, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

func (k Keeper) GetMember(ctx sdk.Context, groupID, memberID uint64) (types.Member, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, false
	}

	member := types.Member{}
	k.cdc.MustUnmarshal(bz, &member)
	return member, true
}

func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID uint64) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

func (k Keeper) GetMembers(ctx sdk.Context, groupID uint64) ([]types.Member, bool) {
	var members []types.Member
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}
	if len(members) == 0 {
		return []types.Member{}, false
	}
	return members, true
}

func (k Keeper) VerifyMember(ctx sdk.Context, groupID, memberID uint64, memberAddress string) bool {
	member, found := k.GetMember(ctx, groupID, memberID)
	if found && member.Signer == memberAddress {
		return true
	}
	return false
}

func (k Keeper) SetRound1Commitments(ctx sdk.Context, groupID uint64, memberID uint64, round1Commitment types.Round1Commitments) {
	ctx.KVStore(k.storeKey).Set(types.Round1CommitmentsMemberStoreKey(groupID, memberID), k.cdc.MustMarshal(&round1Commitment))
}

func (k Keeper) GetRound1Commitments(ctx sdk.Context, groupID uint64, memberID uint64) (types.Round1Commitments, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1CommitmentsMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Commitments{}, false
	}
	var r1c types.Round1Commitments
	k.cdc.MustUnmarshal(bz, &r1c)
	return r1c, true
}

func (k Keeper) getRound1CommitmentsIterator(ctx sdk.Context, groupID uint64) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round1CommitmentsStoreKey(groupID))
}

func (k Keeper) GetRound1CommitmentsCount(ctx sdk.Context, groupID uint64) uint64 {
	var count uint64
	iterator := k.getRound1CommitmentsIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count += 1
	}
	return count
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
