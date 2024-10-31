package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// =====================================
// Process group creation
// =====================================

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
	case types.GROUP_STATUS_ROUND_3:
		// Get members to check malicious
		members := k.MustGetMembers(ctx, group.ID)
		if !types.Members(members).HaveMalicious() {
			group.Status = types.GROUP_STATUS_ACTIVE
			k.SetGroup(ctx, group)

			// Handle the callback when group is successfully created.
			if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
				cb.OnGroupCreationCompleted(ctx, group.ID)
			}

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

			// Handle the callback when fail to create a group.
			if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
				cb.OnGroupCreationFailed(ctx, group.ID)
			}

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

// =====================================
// Process expired group
// =====================================

// HandleExpiredGroups cleans up expired groups and removes them from the store.
func (k Keeper) HandleExpiredGroups(ctx sdk.Context) {
	currentExpiredGroupID := k.GetLastExpiredGroupID(ctx) + 1
	latestGroupID := tss.GroupID(k.GetGroupCount(ctx))

	// Process each group starting from currentGroupID
	groupID := currentExpiredGroupID
	for ; groupID <= latestGroupID; groupID++ {
		group := k.MustGetGroup(ctx, groupID)

		// Check if the group is still within the expiration period
		if group.CreatedHeight+k.GetParams(ctx).CreationPeriod > uint64(ctx.BlockHeight()) {
			break
		}

		// For groups currently undergoing the creation process, set them to be expired.
		if group.Status != types.GROUP_STATUS_ACTIVE && group.Status != types.GROUP_STATUS_FALLEN {
			// Handle the callback before setting group to be expired;
			if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
				cb.OnGroupCreationExpired(ctx, group.ID)
			}

			// Update group status
			group.Status = types.GROUP_STATUS_EXPIRED
			k.SetGroup(ctx, group)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeExpiredGroup,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", group.ID)),
				),
			)
		}

		// Cleanup all interim data associated with the group anyway
		k.DeleteAllDKGInterimData(ctx, groupID)
	}

	// Set the last expired group ID to the previous running group ID
	k.SetLastExpiredGroupID(ctx, groupID-1)
}

// DeleteAllDKGInterimData deletes all DKG interim data for a given groupID.
func (k Keeper) DeleteAllDKGInterimData(ctx sdk.Context, groupID tss.GroupID) {
	// Delete DKG context
	k.DeleteDKGContext(ctx, groupID)

	// Delete Accumulated commits from round 1
	k.DeleteAccumulatedCommits(ctx, groupID)

	// Delete all round 1-3 info
	k.DeleteRound1Infos(ctx, groupID)
	k.DeleteRound2Infos(ctx, groupID)
	k.DeleteConfirmComplains(ctx, groupID)
}

// =====================================
// Pending Process group store
// =====================================

// AddPendingProcessGroup adds a new pending process group to the store.
func (k Keeper) AddPendingProcessGroup(ctx sdk.Context, groupID tss.GroupID) {
	pgs := k.GetPendingProcessGroups(ctx)

	pgs = append(pgs, groupID)
	k.SetPendingProcessGroups(ctx, types.NewPendingProcessGroups(pgs))
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

	var pgs types.PendingProcessGroups
	k.cdc.MustUnmarshal(bz, &pgs)
	return pgs.GroupIDs
}

// =====================================
// Expired Group store
// =====================================

// SetLastExpiredGroupID sets the last expired group ID in the store.
func (k Keeper) SetLastExpiredGroupID(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Set(types.LastExpiredGroupIDStoreKey, sdk.Uint64ToBigEndian(uint64(groupID)))
}

// GetLastExpiredGroupID retrieves the last expired group ID from the store.
func (k Keeper) GetLastExpiredGroupID(ctx sdk.Context) tss.GroupID {
	bz := ctx.KVStore(k.storeKey).Get(types.LastExpiredGroupIDStoreKey)
	return tss.GroupID(sdk.BigEndianToUint64(bz))
}
