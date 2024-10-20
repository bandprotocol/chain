package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// =====================================
// Process fully-submitted group creation message
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

			// Handle the callback when group is ready. this shouldn't return any error.
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

			// Handle the callback when group creation is fallen; this shouldn't return any error.
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

	// Process each group starting from currentGroupID
	for ; currentGroupID <= lastGroupID; currentGroupID++ {
		// Get the group
		group := k.MustGetGroup(ctx, currentGroupID)

		// Check if the group is still within the expiration period
		if group.CreatedHeight+k.GetParams(ctx).CreationPeriod > uint64(ctx.BlockHeight()) {
			break
		}

		// Check group is not active
		if group.Status != types.GROUP_STATUS_ACTIVE && group.Status != types.GROUP_STATUS_FALLEN {
			// Handle the callback before setting group to be expired; this shouldn't return any error.
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

		// Cleanup all interim data associated with the group
		k.DeleteAllDKGInterimData(ctx, currentGroupID)

		// Set the last expired group ID to the current group ID
		k.SetLastExpiredGroupID(ctx, currentGroupID)
	}
}
