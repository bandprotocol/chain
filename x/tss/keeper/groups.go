package keeper

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (k Keeper) CreateGroup(ctx sdk.Context, input types.CreateGroupInput) (*types.CreateGroupResult, error) {
	if k.authority != input.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, input.Authority)
	}

	// Validate group size
	groupSize := uint64(len(input.Members))
	maxGroupSize := k.GetParams(ctx).MaxGroupSize
	if groupSize > maxGroupSize {
		return nil, errors.Wrap(types.ErrGroupSizeTooLarge, fmt.Sprintf("group size exceeds %d", maxGroupSize))
	}

	// Create new group
	fee := input.Fee.Sort()
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: input.Threshold,
		PubKey:    nil,
		Fee:       fee,
		Status:    types.GROUP_STATUS_ROUND_1,
	})

	// Set members
	for i, m := range input.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			return nil, errors.Wrapf(
				types.ErrInvalidAccAddressFormat,
				"invalid account address: %s", err,
			)
		}

		status := k.GetStatus(ctx, address)
		if status.Status != types.MEMBER_STATUS_ACTIVE {
			return nil, types.ErrStatusIsNotActive
		}

		// ID start from 1
		k.SetMember(ctx, types.Member{
			ID:          tss.MemberID(i + 1),
			GroupID:     groupID,
			Address:     m,
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
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", input.Threshold)),
		sdk.NewAttribute(types.AttributeKeyFee, fee.String()),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
	)
	for _, m := range input.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyAddress, m))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.CreateGroupResult{}, nil
}

func (k Keeper) ReplaceGroup(ctx sdk.Context, input types.ReplaceGroupInput) (*types.ReplaceGroupResult, error) {
	if k.authority != input.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, input.Authority)
	}

	address, err := sdk.AccAddressFromBech32(input.Authority)
	if err != nil {
		return nil, errors.Wrapf(
			types.ErrInvalidAccAddressFormat,
			"invalid account address: %s", err,
		)
	}

	// Get new group
	newGroup, err := k.GetGroup(ctx, input.NewGroupID)
	if err != nil {
		return nil, err
	}

	// Verify group status
	if newGroup.Status != types.GROUP_STATUS_ACTIVE {
		return nil, errors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	// Get current group
	currentGroup, err := k.GetGroup(ctx, input.CurrentGroupID)
	if err != nil {
		return nil, err
	}

	// Verify group status
	if currentGroup.Status != types.GROUP_STATUS_ACTIVE {
		return nil, errors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	// Verify whether the group is not in the pending replacement process.
	lastReplacementID := currentGroup.LatestReplacementID
	if lastReplacementID != uint64(0) {
		lastReplacement, err := k.GetReplacement(ctx, lastReplacementID)
		if err != nil {
			panic(err)
		}

		if lastReplacement.Status == types.REPLACEMENT_STATUS_WAITING {
			return nil, errors.Wrap(
				types.ErrRequestReplacementFailed,
				"the group is in the pending replacement process",
			)
		}
	}

	// Request signature
	sid, err := k.HandleReplaceGroupRequestSignature(
		ctx,
		newGroup.PubKey,
		input.CurrentGroupID,
		address,
	)
	if err != nil {
		return nil, err
	}

	nextID := k.GetNextReplacementCount(ctx)
	k.SetReplacement(ctx, types.Replacement{
		ID:             nextID,
		SigningID:      sid,
		CurrentGroupID: input.CurrentGroupID,
		CurrentPubKey:  currentGroup.PubKey,
		NewGroupID:     input.NewGroupID,
		NewPubKey:      newGroup.PubKey,
		ExecTime:       input.ExecTime,
		Status:         types.REPLACEMENT_STATUS_WAITING,
	})

	k.InsertReplacementQueue(ctx, nextID, input.ExecTime)

	// Update latest replacement ID to the group
	currentGroup.LatestReplacementID = nextID
	k.SetGroup(ctx, currentGroup)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(types.AttributeKeyReplacementID, fmt.Sprintf("%d", nextID)),
		),
	)

	return &types.ReplaceGroupResult{}, nil
}
