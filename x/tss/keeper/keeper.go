package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	scopeKeeper capabilitykeeper.ScopedKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	scopeKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		scopeKeeper: scopeKeeper,
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

func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) uint64 {
	id := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(id), k.cdc.MustMarshal(&group))
	return id
}

func (k Keeper) GetGroup(ctx sdk.Context, groupID uint64) types.Group {
	group := types.Group{}
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	k.cdc.MustUnmarshal(bz, &group)
	return group
}

func (k Keeper) SetDKGContext(ctx sdk.Context, groupID uint64, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

func (k Keeper) GetDKGContext(ctx sdk.Context, groupID uint64) []byte {
	return ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
}

func (k Keeper) SetMember(ctx sdk.Context, groupID, memberID uint64, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

func (k Keeper) GetMember(ctx sdk.Context, groupID, memberID uint64, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID uint64) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

func (k Keeper) GetMembers(ctx sdk.Context, groupID uint64) (members []types.Member) {
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}
	return members
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
