package keeper

import (
	"encoding/hex"
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// AddGroup creates a new group in the store and returns the id of the group.
func (k Keeper) AddGroup(
	ctx sdk.Context,
	size uint64,
	threshold uint64,
	moduleOwner string,
) tss.GroupID {
	group := types.NewGroup(
		tss.GroupID(k.GetGroupCount(ctx)+1),
		size,
		threshold,
		nil,
		types.GROUP_STATUS_ROUND_1,
		uint64(ctx.BlockHeight()),
		moduleOwner,
	)

	k.SetGroup(ctx, group)
	k.SetGroupCount(ctx, uint64(group.ID))

	return group.ID
}

// CreateGroup creates a new group with the given members and threshold.
func (k Keeper) CreateGroup(
	ctx sdk.Context,
	members []sdk.AccAddress,
	threshold uint64,
	moduleOwner string,
) (tss.GroupID, error) {
	// Validate group size
	groupSize := uint64(len(members))
	maxGroupSize := k.GetParams(ctx).MaxGroupSize
	if groupSize > maxGroupSize {
		return 0, types.ErrGroupSizeTooLarge.Wrap(fmt.Sprintf("group size exceeds %d", maxGroupSize))
	}

	// add new group
	groupID := k.AddGroup(ctx, groupSize, threshold, moduleOwner)

	// Set members; ID starts from 1
	for i, addr := range members {
		m := types.NewMember(tss.MemberID(i+1), groupID, addr, nil, false, true)
		k.SetMember(ctx, m)
	}

	// Use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(
		uint64(groupID)),
		tss.Hash([]byte(ctx.ChainID())),
		ctx.BlockHeader().LastCommitHash,
	)

	k.SetDKGContext(ctx, groupID, dkgContext)

	event := sdk.NewEvent(
		types.EventTypeCreateGroup,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeySize, fmt.Sprintf("%d", groupSize)),
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", threshold)),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
		sdk.NewAttribute(types.AttributeKeyModuleOwner, moduleOwner),
	)
	for _, m := range members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyAddress, m.String()))
	}
	ctx.EventManager().EmitEvent(event)

	return groupID, nil
}

// GetGroupResponse queries group information from the given id.
func (k Keeper) GetGroupResponse(
	ctx sdk.Context,
	groupID tss.GroupID,
) (*types.GroupResult, error) {
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get group members
	members, err := k.GetGroupMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Ignore error as dkgContext can be deleted
	dkgContext, _ := k.GetDKGContext(ctx, groupID)

	// Get round infos, complaints, and confirms
	round1Infos := k.GetRound1Infos(ctx, groupID)
	round2Infos := k.GetRound2Infos(ctx, groupID)
	complaints := k.GetAllComplainsWithStatus(ctx, groupID)
	confirms := k.GetConfirms(ctx, groupID)

	// Return all the group information
	return &types.GroupResult{
		Group:                group,
		DKGContext:           dkgContext,
		Members:              members,
		Round1Infos:          round1Infos,
		Round2Infos:          round2Infos,
		ComplaintsWithStatus: complaints,
		Confirms:             confirms,
	}, nil
}

// =====================================
// Group store
// =====================================

// GetGroup retrieves a group from the store.
func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, types.ErrGroupNotFound.Wrapf("failed to get group with groupID: %d", groupID)
	}

	var group types.Group
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
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(group.ID), k.cdc.MustMarshal(&group))
}

// GetGroupsIterator gets an iterator all group.
func (k Keeper) GetGroupsIterator(ctx sdk.Context) dbm.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GroupStoreKeyPrefix)
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

// SetGroupCount sets the number of group count to the given value.
func (k Keeper) SetGroupCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetGroupCount returns the current number of all groups ever existed.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey))
}

// =====================================
// DKG store
// =====================================

// SetDKGContext sets DKG context for a group in the store.
func (k Keeper) SetDKGContext(ctx sdk.Context, groupID tss.GroupID, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

// GetDKGContext retrieves DKG context of a group from the store.
func (k Keeper) GetDKGContext(ctx sdk.Context, groupID tss.GroupID) ([]byte, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, types.ErrDKGContextNotFound.Wrapf("failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

// DeleteDKGContext removes the DKG context data of a group from the store.
func (k Keeper) DeleteDKGContext(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.DKGContextStoreKey(groupID))
}
