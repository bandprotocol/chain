package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// CreateGroup creates a new group with the given members and threshold.
func (k Keeper) CreateGroup(ctx sdk.Context, input types.CreateGroupInput) (*types.CreateGroupResult, error) {
	// Create new group
	fee := input.Fee.Sort()
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     uint64(len(input.Members)),
		Threshold: input.Threshold,
		PubKey:    nil,
		Fee:       fee,
		Status:    types.GROUP_STATUS_ROUND_1,
	})

	// Set members
	for i, m := range input.Members {
		k.SetMember(ctx, types.Member{
			ID:          tss.MemberID(i + 1), // ID starts from 1
			GroupID:     groupID,
			Address:     m,
			PubKey:      nil,
			IsMalicious: false,
		})
	}

	// Use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(uint64(groupID)), ctx.BlockHeader().LastCommitHash)
	k.SetDKGContext(ctx, groupID, dkgContext)

	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &types.CreateGroupResult{
		Group:      group,
		DKGContext: dkgContext,
	}, nil
}

// ReplaceGroup creates a new replacement info and put it into queue.
func (k Keeper) ReplaceGroup(ctx sdk.Context, input types.ReplaceGroupInput) (*types.ReplaceGroupResult, error) {
	signingInput := types.CreateSigningInput{
		GroupID:      input.CurrentGroup.ID,
		Message:      append(types.ReplaceGroupMsgPrefix, input.NewGroup.PubKey...),
		IsFeeCharged: input.IsFeeCharged,
		FeeLimit:     sdk.NewCoins(),
		FeePayer:     input.FeePayer,
	}

	signingResult, err := k.CreateSigning(ctx, signingInput)
	if err != nil {
		return nil, err
	}

	nextID := k.GetNextReplacementCount(ctx)
	replacement := types.Replacement{
		ID:             nextID,
		SigningID:      signingResult.Signing.ID,
		CurrentGroupID: input.CurrentGroup.ID,
		CurrentPubKey:  input.CurrentGroup.PubKey,
		NewGroupID:     input.NewGroup.ID,
		NewPubKey:      input.NewGroup.PubKey,
		ExecTime:       input.ExecTime,
		Status:         types.REPLACEMENT_STATUS_WAITING,
	}
	k.SetReplacement(ctx, replacement)

	k.InsertReplacementQueue(ctx, nextID, input.ExecTime)

	// Update latest replacement ID to the group
	input.CurrentGroup.LatestReplacementID = nextID
	k.SetGroup(ctx, input.CurrentGroup)

	return &types.ReplaceGroupResult{Replacement: replacement}, nil
}

// UpdateGroupFee updates the fee of the group.
func (k Keeper) UpdateGroupFee(ctx sdk.Context, input types.UpdateGroupFeeInput) (*types.UpdateGroupFeeResult, error) {
	// Get group
	group, err := k.GetGroup(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	// Set new group fee
	group.Fee = input.Fee.Sort()
	k.SetGroup(ctx, group)

	return &types.UpdateGroupFeeResult{Group: group}, nil
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
