package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

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
