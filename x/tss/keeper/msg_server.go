package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) CreateGroup(goCtx context.Context, req *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupSize := uint64(len(req.Members))

	// Create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Status:    types.ROUND_1,
	})

	// Set members
	for i, m := range req.Members {
		// id start from 1
		k.SetMember(ctx, groupID, tss.MemberID(i+1), types.Member{
			Signer: m,
			PubKey: "",
		})
	}

	// Use LastCommitHash and groupID to hash to dkgContext
	dkgContext := tss.Hash(sdk.Uint64ToBigEndian(uint64(groupID)), ctx.BlockHeader().LastCommitHash)
	k.SetDKGContext(ctx, groupID, dkgContext)

	event := sdk.NewEvent(
		types.EventTypeCreateGroup,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeySize, fmt.Sprintf("%d", groupSize)),
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", req.Threshold)),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
	)
	for _, m := range req.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyMember, m))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateGroupResponse{}, nil
}

func (k Keeper) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 1")
	}

	// Check members
	memberID, err := k.VerifyMember(ctx, groupID, req.Member)
	if err != nil {
		return nil, err
	}

	// Check previous submit
	_, err = k.GetRound1Commitments(ctx, groupID, memberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 1 ")
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	err = tss.VerifyOneTimeSig(memberID, dkgContext, req.OneTimeSig, req.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	err = tss.VerifyA0Sig(memberID, dkgContext, req.A0Sig, tss.PublicKey(req.CoefficientsCommit[0]))
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	round1Commitments := types.Round1Commitments{
		CoefficientsCommit: req.CoefficientsCommit,
		OneTimePubKey:      req.OneTimePubKey,
		A0Sig:              req.A0Sig,
		OneTimeSig:         req.OneTimeSig,
	}
	k.SetRound1Commitments(ctx, groupID, memberID, round1Commitments)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyCoefficientsCommit, round1Commitments.CoefficientsCommit.ToString()),
			sdk.NewAttribute(types.AttributeKeyOneTimePubKey, hex.EncodeToString(round1Commitments.OneTimePubKey)),
			sdk.NewAttribute(types.AttributeKeyA0Sig, hex.EncodeToString(round1Commitments.A0Sig)),
			sdk.NewAttribute(types.AttributeKeyOneTimeSig, hex.EncodeToString(round1Commitments.OneTimeSig)),
		),
	)

	count := k.GetRound1CommitmentsCount(ctx, groupID)
	if count == group.Size_ {
		group.Status = types.ROUND_2
		k.UpdateGroup(ctx, groupID, group)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound1Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	}

	return &types.MsgSubmitDKGRound1Response{}, nil
}

func (k Keeper) SubmitDKGRound2(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound2,
) (*types.MsgSubmitDKGRound2Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_2 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Check members
	memberID, err := k.VerifyMember(ctx, groupID, req.Member)
	if err != nil {
		return nil, err
	}

	// Check previous submit
	_, err = k.GetRound2Share(ctx, groupID, memberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 2")
	}

	// Check round 2 share length
	if uint64(len(req.Round2Share.EncryptedSecretShares)) != group.Size_-1 {
		return nil, sdkerrors.Wrap(types.ErrRound2ShareNotCorrectLength, "number of round 2 share is not correct")
	}

	k.SetRound2Share(ctx, groupID, memberID, *req.Round2Share)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound2,
			sdk.NewAttribute(types.AttributeKeyRound2Share, req.Round2Share.String()),
		),
	)

	count := k.GetRound2SharesCount(ctx, groupID)
	if count == group.Size_ {
		group.Status = types.PENDING
		k.UpdateGroup(ctx, groupID, group)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound2Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
			),
		)
	}

	return &types.MsgSubmitDKGRound2Response{}, nil
}

func (k Keeper) Complain(
	goCtx context.Context,
	req *types.MsgComplain,
) (*types.MsgComplainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.PENDING {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Check members
	_, err = k.VerifyMember(ctx, groupID, req.Member)
	if err != nil {
		return nil, err
	}

	dkgMaliciousIndexes, err := k.GetDKGMaliciousIndexes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Verify complain
	for _, c := range req.Complains {
		// TODO: Verify complain

		if true {
			contains := types.Uint64ArrayContains(dkgMaliciousIndexes.MaliciousIDs, uint64(c.J))
			if contains {
				return nil, sdkerrors.Wrap(types.ErrMemberIsAlreadyMalicious, fmt.Sprintf("member %d is already malicious on this group", c.J))
			}
			dkgMaliciousIndexes.MaliciousIDs = append(dkgMaliciousIndexes.MaliciousIDs, uint64(c.J))
			k.SetDKGMaliciousIndexes(ctx, groupID, dkgMaliciousIndexes)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainsFailed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyComplains, fmt.Sprintf("%+v", c)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		} else {
			contains := types.Uint64ArrayContains(dkgMaliciousIndexes.MaliciousIDs, uint64(c.I))
			if contains {
				return nil, sdkerrors.Wrap(types.ErrMemberIsAlreadyMalicious, fmt.Sprintf("member %d is already malicious on this group", c.I))
			}
			dkgMaliciousIndexes.MaliciousIDs = append(dkgMaliciousIndexes.MaliciousIDs, uint64(c.I))
			k.SetDKGMaliciousIndexes(ctx, groupID, dkgMaliciousIndexes)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainsFailed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyComplains, fmt.Sprintf("%+v", c)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
			return nil, sdkerrors.Wrap(types.ErrComplainsFailed, fmt.Sprintf("failed to complain %+v", c))
		}
	}

	return &types.MsgComplainResponse{}, nil
}

func (k Keeper) Confirm(
	goCtx context.Context,
	req *types.MsgConfirm,
) (*types.MsgConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.PENDING {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Check members
	memberID, err := k.VerifyMember(ctx, groupID, req.Member)
	if err != nil {
		return nil, err
	}

	round1Commitments, err := k.GetRound1Commitments(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	fmt.Println(round1Commitments)

	// TODO: verify OwnPubKeySig

	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	fmt.Println(member)

	dkgMaliciousIndexes, err := k.GetDKGMaliciousIndexes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	pendingRoundNote, err := k.GetPendingRoundNote(ctx, groupID)
	if err != nil {
		return nil, err
	}

	pendingRoundNote.ConfirmationCount += 1

	if pendingRoundNote.ConfirmationCount == group.Size_ && len(dkgMaliciousIndexes.MaliciousIDs) == 0 {
		group.Status = types.ACTIVE

		k.UpdateGroup(ctx, groupID, group)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRound3Success,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
				sdk.NewAttribute(types.AttributeOwnPubKeySig, hex.EncodeToString(req.OwnPubKeySig)),
				sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			),
		)
	}

	// TODO: Remove all interim data associated with this group

	return &types.MsgConfirmResponse{}, nil
}
