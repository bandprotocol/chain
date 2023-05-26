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
			Member:      m,
			PubKey:      tss.PublicKey(nil),
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

	// Verify complain
	var complainsWithStatus []types.ComplainWithStatus
	for _, c := range req.Complains {
		err := k.HandleVerifyComplainSig(ctx, groupID, c)
		if err != nil {
			// handle verify failed
			memberI, err := k.GetMember(ctx, groupID, c.I)
			if err != nil {
				return nil, err
			}
			if memberI.IsMalicious {
				fmt.Printf("member %d is already malicious on this group\n", c.I)
				continue
			}

			// update member status
			memberI.IsMalicious = true
			k.SetMember(ctx, groupID, memberID, memberI)

			// Add complain status
			complainsWithStatus = append(complainsWithStatus, types.ComplainWithStatus{
				Complain:       &c,
				ComplainStatus: types.FAILED,
			})

			// emit complain failed event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainFailed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.I)),
					sdk.NewAttribute(types.AttributeKeyMemberIDJ, fmt.Sprintf("%d", c.J)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeyNonceSym, hex.EncodeToString(c.NonceSym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		} else {
			// handle complains success
			memberJ, err := k.GetMember(ctx, groupID, c.J)
			if err != nil {
				return nil, err
			}
			if memberJ.IsMalicious {
				fmt.Printf("member %d is already malicious on this group\n", c.J)
				continue
			}

			// update member status
			memberJ.IsMalicious = true
			k.SetMember(ctx, groupID, memberID, memberJ)

			// Add complain status
			complainsWithStatus = append(complainsWithStatus, types.ComplainWithStatus{
				Complain:       &c,
				ComplainStatus: types.SUCCESS,
			})

			// Emit complain success event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainSuccess,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyMemberIDI, fmt.Sprintf("%d", c.I)),
					sdk.NewAttribute(types.AttributeKeyMemberIDJ, fmt.Sprintf("%d", c.J)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeyNonceSym, hex.EncodeToString(c.NonceSym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		}

		// Set complain with status
		k.SetComplainsWithStatus(ctx, groupID, memberID, types.ComplainsWithStatus{
			ComplainsWithStatus: complainsWithStatus,
		})

		// Get confirm complain count
		confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)

		// Handle fallen group if everyone sends confirm or complains already.
		if confirmComplainCount == group.Size_ {
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
	err = k.HandleVerifyOwnPubKeySig(ctx, groupID, memberID, req.OwnPubKeySig)
	if err != nil {
		return nil, err
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// TODO: move compute member public key to round 1
	// Compute raw sum commits
	var sumCommits tss.Points
	allRound1Data := k.GetAllRound1Data(ctx, groupID)

	for i := uint64(0); i < group.Threshold; i++ {
		var sumCommitI tss.Points
		for _, r1 := range allRound1Data {
			sumCommitI = append(sumCommitI, r1.CoefficientsCommit[i])
		}
		sumCommit, err := tss.SumPoints(sumCommitI...)
		if err != nil {
			return nil, sdkerrors.Wrapf(
				types.ErrConfirmFailed,
				"can not sum points",
			)
		}
		sumCommits = append(sumCommits, sumCommit)
	}

	// Compute own public key
	ownPubKey, err := tss.ComputeOwnPublicKey(sumCommits, memberID)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrConfirmFailed,
			"compute own public key failed; %s",
			err,
		)
	}

	// Update member
	member.PubKey = ownPubKey
	k.SetMember(ctx, groupID, memberID, member)

	// Get confirm complain count
	confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)

	// Get malicious indexes
	maliciousIndexes, err := k.GetMaliciousIndexes(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Handle fallen group if everyone sends confirm or complains already.
	if confirmComplainCount+1 == group.Size_ {
		if len(maliciousIndexes) == 0 {
			// Handle compute group public key
			groupPubKey, err := k.HandleComputeGroupPublicKey(ctx, groupID)
			if err != nil {
				return nil, err
			}

			// Update group status
			group.Status = types.ACTIVE
			group.PubKey = groupPubKey
			k.UpdateGroup(ctx, groupID, group)

			// emit event round 3 success
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
			k.handleFallenGroup(ctx, groupID, group)

			return nil, sdkerrors.Wrapf(
				types.ErrConfirmFailed,
				"memberIDs: %v is malicious",
				maliciousIndexes,
			)
		}

		// Delete all dkg interim data
		k.DeleteAllDKGInterimData(ctx, groupID, group.Size_)
	}

	// Set Confirm with status
	k.SetConfirm(ctx, groupID, memberID, types.Confirm{
		OwnPubKeySig: req.OwnPubKeySig,
	})

	// emit event confirm success
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmSuccess,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeOwnPubKeySig, hex.EncodeToString(req.OwnPubKeySig)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
		),
	)

	return &types.MsgConfirmResponse{}, nil
}

// Check already confirm or complain
func (k Keeper) checkConfirmOrComplain(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	_, err := k.GetConfirm(ctx, groupID, memberID)
	if err == nil {
		return sdkerrors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send confirm message",
			memberID,
		)
	}
	_, err = k.GetComplainsWithStatus(ctx, groupID, memberID)
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
