package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

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

// CreateNewGroup creates a new group in the store and returns the id of the group.
func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) tss.GroupID {
	group.ID = k.GetNextGroupID(ctx)
	group.CreatedHeight = uint64(ctx.BlockHeight())
	k.SetGroup(ctx, group)

	return group.ID
}

// GetGroup retrieves a group from the store.
func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, types.ErrGroupNotFound.Wrapf("failed to get group with groupID: %d", groupID)
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
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(group.ID), k.cdc.MustMarshal(&group))
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
		return nil, types.ErrDKGContextNotFound.Wrapf("failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

// DeleteDKGContext removes the DKG context data of a group from the store.
func (k Keeper) DeleteDKGContext(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.DKGContextStoreKey(groupID))
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

	// Create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:       groupSize,
		Threshold:   threshold,
		PubKey:      nil,
		Status:      types.GROUP_STATUS_ROUND_1,
		ModuleOwner: moduleOwner,
	})

	// Set members
	for i, m := range members {
		k.SetMember(ctx, types.Member{
			ID:          tss.MemberID(i + 1), // ID starts from 1
			GroupID:     groupID,
			Address:     m.String(),
			PubKey:      nil,
			IsMalicious: false,
			IsActive:    true,
		})
	}

	// Use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(uint64(groupID)), ctx.BlockHeader().LastCommitHash)
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

// GetPenalizedMembersExpiredGroup gets the list of members who should be penalized due to not
// participating in group creation.
func (k Keeper) GetPenalizedMembersExpiredGroup(ctx sdk.Context, group types.Group) ([]sdk.AccAddress, error) {
	members, err := k.GetGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	var penalizedMembers []sdk.AccAddress
	for _, m := range members {
		address := sdk.MustAccAddressFromBech32(m.Address)

		// query if the member send a message, if not then penalize.
		switch group.Status {
		case types.GROUP_STATUS_ROUND_1:
			_, err := k.GetRound1Info(ctx, group.ID, m.ID)
			if err != nil {
				penalizedMembers = append(penalizedMembers, address)
			}
		case types.GROUP_STATUS_ROUND_2:
			_, err := k.GetRound2Info(ctx, group.ID, m.ID)
			if err != nil {
				penalizedMembers = append(penalizedMembers, address)
			}
		case types.GROUP_STATUS_ROUND_3:
			err := k.checkConfirmOrComplain(ctx, group.ID, m.ID)
			if err != nil {
				penalizedMembers = append(penalizedMembers, address)
			}
		default:
		}
	}
	return penalizedMembers, nil
}
