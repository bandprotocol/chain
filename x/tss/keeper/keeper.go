package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc               codec.BinaryCodec
	storeKey          storetypes.StoreKey
	paramSpace        paramtypes.Subspace
	authzKeeper       types.AuthzKeeper
	rollingseedKeeper types.RollingseedKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authzKeeper types.AuthzKeeper,
	rollingseedKeeper types.RollingseedKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		paramSpace:        paramSpace,
		authzKeeper:       authzKeeper,
		rollingseedKeeper: rollingseedKeeper,
	}
}

// SetGroupCount sets the number of group count to the given value.
func (k Keeper) SetGroupCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetGroupCount returns the current number of all groups ever existed.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey))
}

// GetNextGroupID increments the group count and returns the current number of groups.
func (k Keeper) GetNextGroupID(ctx sdk.Context) tss.GroupID {
	groupNumber := k.GetGroupCount(ctx)
	k.SetGroupCount(ctx, groupNumber+1)
	return tss.GroupID(groupNumber + 1)
}

// IsGrantee checks if the granter granted permissions to the grantee.
func (k Keeper) IsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
	for _, msg := range types.GetMsgGrants() {
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

// CreateNewGroup creates a new group in the store and returns the id of the group.
func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) tss.GroupID {
	groupID := k.GetNextGroupID(ctx)
	group.GroupID = groupID
	group.CreatedHeight = ctx.BlockHeader().Height
	k.SetGroup(ctx, group)

	return groupID
}

// GetGroup retrieves a group from the store.
func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get group with groupID: %d", groupID)
	}

	group := types.Group{}
	k.cdc.MustUnmarshal(bz, &group)
	return group, nil
}

// MustGetGroup returns the group for the given ID. Panics error if not exists.
func (k Keeper) MustGetGroup(ctx sdk.Context, groupID tss.GroupID) types.Group {
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		panic(err)
	}
	return group
}

// SetGroup set a group in the store.
func (k Keeper) SetGroup(ctx sdk.Context, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(group.GroupID), k.cdc.MustMarshal(&group))
}

// GetGroupsIterator gets an iterator all group.
func (k Keeper) GetGroupsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GroupStoreKeyPrefix)
}

// GetGroups retrieves all group of the store.
func (k Keeper) GetGroups(ctx sdk.Context) []types.Group {
	var groups []types.Group
	iterator := k.GetGroupsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var group types.Group
		k.cdc.MustUnmarshal(iterator.Value(), &group)
		groups = append(groups, group)
	}
	return groups
}

// DeleteGroup removes the group from the store.
func (k Keeper) DeleteGroup(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.GroupStoreKey(groupID))
}

// SetDKGContext sets DKG context for a group in the store.
func (k Keeper) SetDKGContext(ctx sdk.Context, groupID tss.GroupID, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

// GetDKGContext retrieves DKG context of a group from the store.
func (k Keeper) GetDKGContext(ctx sdk.Context, groupID tss.GroupID) ([]byte, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDKGContextNotFound, "failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

// DeleteDKGContext removes the DKG context data of a group from the store.
func (k Keeper) DeleteDKGContext(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.DKGContextStoreKey(groupID))
}

// SetMember sets a member of a group in the store.
func (k Keeper) SetMember(ctx sdk.Context, groupID tss.GroupID, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, member.MemberID), k.cdc.MustMarshal(&member))
}

// GetMemberByAddress function retrieves a member of a group from the store by using address.
func (k Keeper) GetMemberByAddress(ctx sdk.Context, groupID tss.GroupID, address string) (types.Member, error) {
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return types.Member{}, err
	}

	for _, member := range members {
		if member.Verify(address) {
			return member, nil
		}
	}

	return types.Member{}, sdkerrors.Wrapf(
		types.ErrMemberNotFound,
		"failed to get member with groupID: %d and address: ",
		groupID,
		address,
	)
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

// GetMembersIterator gets an iterator over all members of a group.
func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

// GetMembers retrieves all members of a group from the store.
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
		return nil, sdkerrors.Wrapf(types.ErrMemberNotFound, "failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// GetActiveMembers retrieves all active members of a group from the store.
func (k Keeper) GetActiveMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var members []types.Member
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)

		if member.IsActive {
			members = append(members, member)
		}
	}

	// Filter members that have DE left
	filteredMembers, err := k.FilterMembersHaveDE(ctx, members)
	if err != nil {
		return nil, err
	}

	if len(filteredMembers) == 0 {
		return nil, sdkerrors.Wrapf(types.ErrNoActiveMember, "no active member in groupID: %d", groupID)
	}
	return filteredMembers, nil
}

// SetLastExpiredGroupID sets the last expired group ID in the store.
func (k Keeper) SetLastExpiredGroupID(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Set(types.LastExpiredGroupIDStoreKey, sdk.Uint64ToBigEndian(uint64(groupID)))
}

// GetLastExpiredGroupID retrieves the last expired group ID from the store.
func (k Keeper) GetLastExpiredGroupID(ctx sdk.Context) tss.GroupID {
	bz := ctx.KVStore(k.storeKey).Get(types.LastExpiredGroupIDStoreKey)
	return tss.GroupID(sdk.BigEndianToUint64(bz))
}

// ProcessExpiredGroups cleans up expired groups and removes them from the store.
func (k Keeper) ProcessExpiredGroups(ctx sdk.Context) {
	// Get the current group ID to start processing from
	currentGroupID := k.GetLastExpiredGroupID(ctx) + 1

	// Get the last group ID in the store
	lastGroupID := tss.GroupID(k.GetGroupCount(ctx))

	// Get the group signature creating period
	creatingPeriod := k.CreatingPeriod(ctx)

	// Process each group starting from currentGroupID
	for ; currentGroupID <= lastGroupID; currentGroupID++ {
		// Get the group
		group := k.MustGetGroup(ctx, currentGroupID)

		// Check if the group is still within the expiration period
		if group.CreatedHeight+creatingPeriod > ctx.BlockHeight() {
			break
		}

		// Check group is not active
		if group.Status != types.GROUP_STATUS_ACTIVE {
			// Update group status
			group.Status = types.GROUP_STATUS_EXPIRED
			k.SetGroup(ctx, group)
		}

		// Cleanup all interim data associated with the group
		k.DeleteAllDKGInterimData(ctx, currentGroupID)

		// Set the last expired group ID to the current group ID
		k.SetLastExpiredGroupID(ctx, currentGroupID)
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
