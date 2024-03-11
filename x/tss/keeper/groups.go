package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// CreateGroup creates a new group with the given members and threshold.
func (k Keeper) CreateGroup(
	ctx sdk.Context,
	members []sdk.AccAddress,
	threshold uint64,
	fee sdk.Coins,
) (tss.GroupID, error) {
	// Validate group size
	groupSize := uint64(len(members))
	maxGroupSize := k.GetParams(ctx).MaxGroupSize
	if groupSize > maxGroupSize {
		return 0, types.ErrGroupSizeTooLarge.Wrap(fmt.Sprintf("group size exceeds %d", maxGroupSize))
	}

	// Create new group
	sortedFee := fee.Sort()
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: threshold,
		PubKey:    nil,
		Fee:       sortedFee,
		Status:    types.GROUP_STATUS_ROUND_1,
	})

	// Set members
	for i, m := range members {
		k.SetMember(ctx, types.Member{
			ID:          tss.MemberID(i + 1), // ID starts from 1
			GroupID:     groupID,
			Address:     m.String(),
			PubKey:      nil,
			IsMalicious: false,
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
		sdk.NewAttribute(types.AttributeKeyFee, fee.String()),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
	)
	for _, m := range members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyAddress, m.String()))
	}
	ctx.EventManager().EmitEvent(event)

	return groupID, nil
}

// ReplaceGroup creates a new replacement info and put it into queue.
func (k Keeper) ReplaceGroup(
	ctx sdk.Context,
	currentGroupID tss.GroupID,
	newGroupID tss.GroupID,
	execTime time.Time,
	feePayer sdk.AccAddress,
	fee sdk.Coins,
) (uint64, error) {
	// Check if NewGroupID and CurrentGroupID are active
	newGroup, err := k.GetActiveGroup(ctx, newGroupID)
	if err != nil {
		return 0, err
	}

	currentGroup, err := k.GetActiveGroup(ctx, currentGroupID)
	if err != nil {
		return 0, err
	}

	// Verify whether the group is not in the pending replacement process.
	lastReplacementID := currentGroup.LatestReplacementID
	if lastReplacementID != uint64(0) {
		lastReplacement, err := k.GetReplacement(ctx, lastReplacementID)
		if err != nil {
			return 0, err
		}

		if lastReplacement.Status == types.REPLACEMENT_STATUS_WAITING {
			return 0, types.ErrRequestReplacementFailed.Wrap(
				"the group is in the pending replacement process",
			)
		}
	}

	msg := append(types.ReplaceGroupMsgPrefix, newGroup.PubKey...)
	signing, err := k.CreateSigning(ctx, currentGroup, msg, fee, feePayer)
	if err != nil {
		return 0, err
	}

	nextID := k.GetNextReplacementCount(ctx)
	replacement := types.Replacement{
		ID:             nextID,
		SigningID:      signing.ID,
		CurrentGroupID: currentGroup.ID,
		CurrentPubKey:  currentGroup.PubKey,
		NewGroupID:     newGroup.ID,
		NewPubKey:      newGroup.PubKey,
		ExecTime:       execTime,
		Status:         types.REPLACEMENT_STATUS_WAITING,
	}
	k.SetReplacement(ctx, replacement)

	k.InsertReplacementQueue(ctx, nextID, execTime)

	// Update latest replacement ID to the group
	currentGroup.LatestReplacementID = nextID
	k.SetGroup(ctx, currentGroup)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(types.AttributeKeyReplacementID, fmt.Sprintf("%d", replacement.ID)),
		),
	)

	return replacement.ID, nil
}

// UpdateGroupFee updates the fee of the group.
func (k Keeper) UpdateGroupFee(
	ctx sdk.Context,
	groupID tss.GroupID,
	fee sdk.Coins,
) (*types.Group, error) {
	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Set new group fee
	group.Fee = fee.Sort()
	k.SetGroup(ctx, group)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateGroupFee,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyFee, group.Fee.String()),
		),
	)

	return &group, nil
}

// GetActiveGroup returns the active group with the given groupID. If the group is not active,
// return error.
func (k Keeper) GetActiveGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return types.Group{}, err
	}

	if group.Status != types.GROUP_STATUS_ACTIVE {
		return types.Group{}, types.ErrGroupIsNotActive.Wrap("group status is not active")
	}

	return group, nil
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
