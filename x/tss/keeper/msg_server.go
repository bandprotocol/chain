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

// CreateGroup handles the request to create a new group.
// It first unwraps the Go context into an SDK context, then creates a new group with the given members.
// Afterwards, it sets each member into the KVStore, and hashes the groupID with the LastCommitHash from the block header to create the DKG context.
// Finally, it emits an event for the group creation.
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
			Member: m,
			PubKey: tss.PublicKey(nil),
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

// SubmitDKGRound1 handles the submission of round 1 in the DKG process.
// After unwrapping the context, it first checks the status of the group, and whether the member is valid and has not submitted before.
// Then, it retrieves the DKG context for the group and verifies the one-time signature and A0 signature.
// If all checks pass, it saves the round 1 data into the KVStore and emits an event for the submission.
// If all members have submitted their round 1 data, it updates the status of the group to round 2 and emits an event for the completion of round 1.
func (k Keeper) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round1Data.MemberID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 1")
	}

	// Verify member
	isMember := k.VerifyMember(ctx, groupID, memberID, req.Member)
	if !isMember {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not in this group",
			memberID,
			req.Member,
		)
	}

	// Check previous submit
	_, err = k.GetRound1Data(ctx, groupID, req.Round1Data.MemberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 1 ")
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	err = tss.VerifyOneTimeSig(memberID, dkgContext, req.Round1Data.OneTimeSig, req.Round1Data.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	err = tss.VerifyA0Sig(
		memberID,
		dkgContext,
		req.Round1Data.A0Sig,
		tss.PublicKey(req.Round1Data.CoefficientsCommit[0]),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	k.SetRound1Data(ctx, groupID, req.Round1Data)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			sdk.NewAttribute(
				types.AttributeKeyRound1Data,
				hex.EncodeToString(k.cdc.MustMarshal(&req.Round1Data)),
			),
		),
	)

	count := k.GetRound1DataCount(ctx, groupID)
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

// SubmitDKGRound2 is responsible for handling the submission of DKG (Distributed Key Generation) round2.
// It verifies the group status, member authorization, previous submission, and the correctness of encrypted secret shares length.
// It sets the round2 data, emits events, and updates the group status.
func (k Keeper) SubmitDKGRound2(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound2,
) (*types.MsgSubmitDKGRound2Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round2Data.MemberID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_2 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Verify member
	isMember := k.VerifyMember(ctx, groupID, memberID, req.Member)
	if !isMember {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not in this group",
			memberID,
			req.Member,
		)
	}

	// Check previous submit
	_, err = k.GetRound2Data(ctx, groupID, memberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 2")
	}

	// Check encrypted secret shares length
	if uint64(len(req.Round2Data.EncryptedSecretShares)) != group.Size_-1 {
		return nil, sdkerrors.Wrap(
			types.ErrEncryptedSecretSharesNotCorrectLength,
			"number of encrypted secret shares is not correct",
		)
	}

	k.SetRound2Data(ctx, groupID, req.Round2Data)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound2,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			sdk.NewAttribute(types.AttributeKeyRound2Data, hex.EncodeToString(k.cdc.MustMarshal(&req.Round2Data))),
		),
	)

	count := k.GetRound2DataCount(ctx, groupID)
	if count == group.Size_ {
		group.Status = types.ROUND_3
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
	memberID := req.MemberID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.ROUND_3 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Verify member
	isMember := k.VerifyMember(ctx, groupID, memberID, req.Member)
	if !isMember {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not in this group",
			memberID,
			req.Member,
		)
	}

	// Check already confirm or complain
	err = k.checkConfirmOrComplain(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Get dkg malicious indexes
	dkgMaliciousIndexes, err := k.GetDKGMaliciousIndexes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Verify complain
	for _, c := range req.Complains {
		err := k.VerifyComplainSig(ctx, groupID, *c)
		if err != nil {
			// handle verify failed
			contains := types.Uint64ArrayContains(dkgMaliciousIndexes.MaliciousIDs, uint64(c.I))
			if contains {
				return nil, sdkerrors.Wrap(
					types.ErrMemberIsAlreadyMalicious,
					fmt.Sprintf("member %d is already malicious on this group", c.I),
				)
			}
			dkgMaliciousIndexes.MaliciousIDs = append(dkgMaliciousIndexes.MaliciousIDs, uint64(c.I))
			k.SetDKGMaliciousIndexes(ctx, groupID, dkgMaliciousIndexes)

			// emit complain failed event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainFailed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.I)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.J)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeyNonceSym, hex.EncodeToString(c.Noncesym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		} else {
			// handle complains success
			contains := types.Uint64ArrayContains(dkgMaliciousIndexes.MaliciousIDs, uint64(c.J))
			if contains {
				return nil, sdkerrors.Wrap(
					types.ErrMemberIsAlreadyMalicious,
					fmt.Sprintf("member %d is already malicious on this group", c.J),
				)
			}
			dkgMaliciousIndexes.MaliciousIDs = append(dkgMaliciousIndexes.MaliciousIDs, uint64(c.J))
			k.SetDKGMaliciousIndexes(ctx, groupID, dkgMaliciousIndexes)

			// emit complain success event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainSuccess,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.I)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.J)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeyNonceSym, hex.EncodeToString(c.Noncesym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		}

		// Get round 3 note.
		round3Note, err := k.GetRound3Note(ctx, groupID)
		if err != nil {
			return nil, err
		}

		round3Note.ConfirmComplainCount += 1
		// Handle fallen group if everyone sends confirm or complains already.
		if round3Note.ConfirmComplainCount == group.Size_ {
			k.handleFallenGroup(ctx, groupID, group)
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
	memberID := req.MemberID

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_3 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Verify member
	isMember := k.VerifyMember(ctx, groupID, memberID, req.Member)
	if !isMember {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not in this group",
			memberID,
			req.Member,
		)
	}

	// Check already confirm or complain
	err = k.checkConfirmOrComplain(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify OwnPubKeySig
	err = k.VerifyOwnPubKeySig(ctx, groupID, memberID, req.OwnPubKeySig)
	if err != nil {
		return nil, err
	}

	// Get dkg malicious indexes
	dkgMaliciousIndexes, err := k.GetDKGMaliciousIndexes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	round3Note, err := k.GetRound3Note(ctx, groupID)
	if err != nil {
		return nil, err
	}

	round3Note.ConfirmComplainCount += 1
	if round3Note.ConfirmComplainCount == group.Size_ {
		// Handle active group
		if len(dkgMaliciousIndexes.MaliciousIDs) == 0 {
			// TODO: Compute final group public key

			groupPubKey, err := tss.ComputeGroupPublicKey(tss.Points{})
			if err != nil {
				return nil, sdkerrors.Wrapf(
					types.ErrConfirmFailed,
					"failed to compute group public key; %s",
					err,
				)
			}

			group.Status = types.ACTIVE
			group.PubKey = groupPubKey

			k.UpdateGroup(ctx, groupID, group)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRound3Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeOwnPubKeySig, hex.EncodeToString(req.OwnPubKeySig)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		} else {
			// Handle fallen group if someone in this group is malicious.
			if round3Note.ConfirmComplainCount == group.Size_ {
				k.handleFallenGroup(ctx, groupID, group)
			}
		}
	}

	// TODO: Remove all interim data associated with round 3

	return &types.MsgConfirmResponse{}, nil
}

// Check already confirm or complain
func (k Keeper) checkConfirmOrComplain(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	_, err := k.GetConfirmation(ctx, groupID, memberID)
	if err == nil {
		return sdkerrors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send confirm message",
			memberID,
		)
	}
	_, err = k.GetComplain(ctx, groupID, memberID)
	if err == nil {
		return sdkerrors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send complain message",
			memberID,
		)
	}
	return nil
}

// HandleFallenGroup updates the status of a group and emit event.
func (k Keeper) handleFallenGroup(
	ctx sdk.Context,
	groupID tss.GroupID,
	group types.Group,
) {
	group.Status = types.FALLEN

	k.UpdateGroup(ctx, groupID, group)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRound3Failed,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
		),
	)
}
