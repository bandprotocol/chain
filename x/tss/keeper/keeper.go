package keeper

import (
	"bytes"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         storetypes.StoreKey
	paramSpace       paramtypes.Subspace
	feeCollectorName string

	authzKeeper       types.AuthzKeeper
	rollingseedKeeper types.RollingseedKeeper
	authKeeper        types.AccountKeeper
	bankKeeper        types.BankKeeper
	stakingKeeper     types.StakingKeeper
	distrKeeper       types.DistrKeeper

	router    types.Router
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	feeCollectorName string,
	authzKeeper types.AuthzKeeper,
	rollingseedKeeper types.RollingseedKeeper,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistrKeeper,
	rtr types.Router,
	authority string,
) Keeper {
	// ensure TSS module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		paramSpace:        paramSpace,
		feeCollectorName:  feeCollectorName,
		authzKeeper:       authzKeeper,
		rollingseedKeeper: rollingseedKeeper,
		authKeeper:        authKeeper,
		bankKeeper:        bankKeeper,
		stakingKeeper:     stakingKeeper,
		distrKeeper:       distrKeeper,
		router:            rtr,
		authority:         authority,
	}
}

// GetAuthority returns the x/tss module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetTSSAccount returns the TSS ModuleAccount
func (k Keeper) GetTSSAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
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

// CheckIsGrantee checks if the granter granted permissions to the grantee.
func (k Keeper) CheckIsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
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
		return types.Group{}, errors.Wrapf(types.ErrGroupNotFound, "failed to get group with groupID: %d", groupID)
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
		return nil, errors.Wrapf(types.ErrDKGContextNotFound, "failed to get dkg-context with groupID: %d", groupID)
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

// SetMembers sets members of a group in the store.
func (k Keeper) SetMembers(ctx sdk.Context, groupID tss.GroupID, members []types.Member) {
	for _, member := range members {
		k.SetMember(ctx, groupID, member)
	}
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

	return types.Member{}, errors.Wrapf(
		types.ErrMemberNotFound,
		"failed to get member with groupID: %d and address: %s",
		groupID,
		address,
	)
}

// GetMember function retrieves a member of a group from the store.
func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, errors.Wrapf(
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

// MustGetMember returns the member for the given groupID and memberID. Panics error if not exists.
func (k Keeper) MustGetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) types.Member {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		panic(err)
	}
	return member
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
		return nil, errors.Wrapf(types.ErrMemberNotFound, "failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// MustGetMembers retrieves all members of a group from the store. Panics error if not exists.
func (k Keeper) MustGetMembers(ctx sdk.Context, groupID tss.GroupID) []types.Member {
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		panic(err)
	}
	return members
}

// GetActiveMembers retrieves all active members of a group from the store.
func (k Keeper) GetActiveMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var members []types.Member
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)

		address := sdk.MustAccAddressFromBech32(member.Address)
		status := k.GetStatus(ctx, address)
		if status.Status == types.MEMBER_STATUS_ACTIVE {
			members = append(members, member)
		}
	}

	// Filter members that have DE left
	filteredMembers, err := k.FilterMembersHaveDE(ctx, members)
	if err != nil {
		return nil, err
	}

	if len(filteredMembers) == 0 {
		return nil, errors.Wrapf(types.ErrNoActiveMember, "no active member in groupID: %d", groupID)
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

// HandleExpiredGroups cleans up expired groups and removes them from the store.
func (k Keeper) HandleExpiredGroups(ctx sdk.Context) {
	// Get the current group ID to start processing from
	currentGroupID := k.GetLastExpiredGroupID(ctx) + 1

	// Get the last group ID in the store
	lastGroupID := tss.GroupID(k.GetGroupCount(ctx))

	// Get the group signature creating period
	creatingPeriod := k.GetParams(ctx).CreatingPeriod

	// Process each group starting from currentGroupID
	for ; currentGroupID <= lastGroupID; currentGroupID++ {
		// Get the group
		group := k.MustGetGroup(ctx, currentGroupID)

		// Check if the group is still within the expiration period
		if group.CreatedHeight+creatingPeriod > ctx.BlockHeight() {
			break
		}

		// Check group is not active
		if group.Status != types.GROUP_STATUS_ACTIVE && group.Status != types.GROUP_STATUS_FALLEN {
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

// SetStatus sets a status of a member of the group in the store.
func (k Keeper) SetStatus(ctx sdk.Context, status types.Status) {
	address := sdk.MustAccAddressFromBech32(status.Address)
	ctx.KVStore(k.storeKey).Set(types.StatusStoreKey(address), k.cdc.MustMarshal(&status))
}

// GetStatusesIterator gets an iterator all statuses of address.
func (k Keeper) GetStatusesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StatusStoreKeyPrefix)
}

// GetStatus retrieves a status of the address.
func (k Keeper) GetStatus(ctx sdk.Context, address sdk.AccAddress) types.Status {
	bz := ctx.KVStore(k.storeKey).Get(types.StatusStoreKey(address))
	if bz == nil {
		return types.Status{
			Address: address.String(),
			Status:  types.MEMBER_STATUS_UNSPECIFIED,
			Since:   time.Time{},
		}
	}

	status := types.Status{}
	k.cdc.MustUnmarshal(bz, &status)
	return status
}

// DeleteStatus removes the status of the address of the group
func (k Keeper) DeleteStatus(ctx sdk.Context, address sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(types.StatusStoreKey(address))
}

// AddPendingProcessGroup adds a new pending process group to the store.
func (k Keeper) AddPendingProcessGroup(ctx sdk.Context, groupID tss.GroupID) {
	pgs := k.GetPendingProcessGroups(ctx)
	pgs = append(pgs, groupID)
	k.SetPendingProcessGroups(ctx, types.PendingProcessGroups{
		GroupIDs: pgs,
	})
}

// SetPendingProcessGroups sets the given pending process groups in the store.
func (k Keeper) SetPendingProcessGroups(ctx sdk.Context, pgs types.PendingProcessGroups) {
	ctx.KVStore(k.storeKey).Set(types.PendingProcessGroupsStoreKey, k.cdc.MustMarshal(&pgs))
}

// GetPendingProcessGroups retrieves the list of pending process groups from the store.
// It returns an empty list if the key does not exist in the store.
func (k Keeper) GetPendingProcessGroups(ctx sdk.Context) []tss.GroupID {
	bz := ctx.KVStore(k.storeKey).Get(types.PendingProcessGroupsStoreKey)
	if len(bz) == 0 {
		// Return an empty list if the key does not exist in the store.
		return []tss.GroupID{}
	}
	pgs := types.PendingProcessGroups{}
	k.cdc.MustUnmarshal(bz, &pgs)
	return pgs.GroupIDs
}

// HandleProcessGroup handles the pending process group based on its status.
// It updates the group status and emits appropriate events.
func (k Keeper) HandleProcessGroup(ctx sdk.Context, groupID tss.GroupID) {
	group := k.MustGetGroup(ctx, groupID)
	switch group.Status {
	case types.GROUP_STATUS_ROUND_1:
		group.Status = types.GROUP_STATUS_ROUND_2
		group.PubKey = k.GetAccumulatedCommit(ctx, groupID, 0)
		k.SetGroup(ctx, group)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound1Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	case types.GROUP_STATUS_ROUND_2:
		group.Status = types.GROUP_STATUS_ROUND_3
		k.SetGroup(ctx, group)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound2Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	case types.GROUP_STATUS_FALLEN:
		group.Status = types.GROUP_STATUS_FALLEN
		k.SetGroup(ctx, group)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound3Failed,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	case types.GROUP_STATUS_ROUND_3:
		// Get members to check malicious
		members := k.MustGetMembers(ctx, group.GroupID)
		if !types.Members(members).HaveMalicious() {
			group.Status = types.GROUP_STATUS_ACTIVE
			k.SetGroup(ctx, group)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRound3Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
				),
			)
		} else {
			group.Status = types.GROUP_STATUS_FALLEN
			k.SetGroup(ctx, group)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRound3Failed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
				),
			)
		}
	}
}

// SetReplacementCount sets the number of replacement group count to the given value.
func (k Keeper) SetReplacementCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.ReplacementCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetReplacementCount returns the current number of all replacements ever existed.
func (k Keeper) GetReplacementCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.ReplacementCountStoreKey))
}

// GetNextReplacementCount increments the replacement count and returns the current number of groups.
func (k Keeper) GetNextReplacementCount(ctx sdk.Context) uint64 {
	replacementNumber := k.GetReplacementCount(ctx)
	k.SetReplacementCount(ctx, replacementNumber+1)
	return replacementNumber + 1
}

// GetReplacement gets a replacement of store by ReplacementID.
func (k Keeper) GetReplacement(ctx sdk.Context, replacementID uint64) (types.Replacement, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ReplacementKey(replacementID))
	if bz == nil {
		return types.Replacement{}, errors.Wrapf(types.ErrReplacementNotFound, "failed to get replacement group with replacement ID: %d", replacementID)
	}

	replacement := types.Replacement{}
	k.cdc.MustUnmarshal(bz, &replacement)
	return replacement, nil
}

// MustGetReplacement gets a replacement of store by ReplacementID. Panics error if not exists.
func (k Keeper) MustGetReplacement(ctx sdk.Context, replacementID uint64) types.Replacement {
	replacement, err := k.GetReplacement(ctx, replacementID)
	if err != nil {
		panic(err)
	}
	return replacement
}

// SetReplacement sets a replacement to store.
func (k Keeper) SetReplacement(ctx sdk.Context, replacement types.Replacement) {
	ctx.KVStore(k.storeKey).Set(types.ReplacementKey(replacement.ID), k.cdc.MustMarshal(&replacement))
}

// InsertReplacementQueue inserts a replacementID into the replacement queue at endTime
func (k Keeper) InsertReplacementQueue(ctx sdk.Context, replacementID uint64, endTime time.Time) {
	ctx.KVStore(k.storeKey).Set(types.ReplacementQueueKey(replacementID, endTime), sdk.Uint64ToBigEndian(replacementID))
}

// RemoveFromReplacementQueue removes a replacementID from the replacement queue.
func (keeper Keeper) RemoveFromReplacementQueue(ctx sdk.Context, replacementID uint64, endTime time.Time) {
	ctx.KVStore(keeper.storeKey).Delete(types.ReplacementQueueKey(replacementID, endTime))
}

// IterateReplacementQueue iterates over the replacements in the active proposal replacement group queue.
// and performs a callback function
func (k Keeper) IterateReplacementQueue(ctx sdk.Context, endTime time.Time, cb func(replacement types.Replacement) (stop bool)) {
	iterator := k.ReplacementQueueIterator(ctx, endTime)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		replacementID, _ := types.SplitReplacementQueueKey(iterator.Key())
		replacement, err := k.GetReplacement(ctx, replacementID)
		if err != nil {
			panic(fmt.Sprintf("replacement group ID %d does not exist", replacementID))
		}

		if cb(replacement) {
			break
		}
	}
}

// ReplacementQueueIterator returns an sdk.Iterator for all the replacements in the replacement group Queue that expire by endTime
func (keeper Keeper) ReplacementQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(types.ReplacementQueuePrefix, sdk.PrefixEndBytes(types.ReplacementQueueByTimeKey(endTime)))
}

// HandleReplaceGroup updates the group information after a successful signing process.
func (k Keeper) HandleReplaceGroup(ctx sdk.Context, replacement types.Replacement) {
	// Retrieve information about signing.
	signing := k.MustGetSigning(ctx, replacement.SigningID)

	// If the signing process is unsuccessful, update the replacement status to failed.
	if signing.Status != types.SIGNING_STATUS_SUCCESS {
		replacement.Status = types.REPLACEMENT_STATUS_FALLEN
		k.SetReplacement(ctx, replacement)
		return
	}

	// Retrieve information about group.
	fromGroup := k.MustGetGroup(ctx, replacement.FromGroupID)
	toGroup := k.MustGetGroup(ctx, replacement.ToGroupID)

	// If the group's public key is changed, update the replacement status to failed.
	if !bytes.Equal(fromGroup.PubKey, replacement.FromPubKey) || !bytes.Equal(toGroup.PubKey, replacement.ToPubKey) {
		replacement.Status = types.REPLACEMENT_STATUS_FALLEN
		k.SetReplacement(ctx, replacement)
		return
	}

	// Replace group data
	tempGroup := fromGroup
	tempGroup.GroupID = toGroup.GroupID
	tempGroup.CreatedHeight = toGroup.CreatedHeight
	tempGroup.LatestReplacementID = toGroup.LatestReplacementID

	// Set group with new data
	k.SetGroup(ctx, tempGroup)

	// Set members with new data
	members, err := k.GetMembers(ctx, replacement.FromGroupID)
	if err != nil {
		return
	}
	k.SetMembers(ctx, replacement.ToGroupID, members)

	// Update replacement group status to success
	replacement.Status = types.REPLACEMENT_STATUS_SUCCESS
	k.SetReplacement(ctx, replacement)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplaceSuccess,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", replacement.SigningID)),
			sdk.NewAttribute(types.AttributeKeyFromGroupID, fmt.Sprintf("%d", replacement.FromGroupID)),
			sdk.NewAttribute(types.AttributeKeyToGroupID, fmt.Sprintf("%d", replacement.ToGroupID)),
		),
	)
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
