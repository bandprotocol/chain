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

	// Check coefficients commit length
	if uint64(len(req.Round1Data.CoefficientsCommit)) != group.Threshold {
		return nil, sdkerrors.Wrap(
			types.ErrCommitsNotCorrectLength,
			"number of coefficients commit is not correct",
		)
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	// Verify one time signature
	err = tss.VerifyOneTimeSig(memberID, dkgContext, req.Round1Data.OneTimeSig, req.Round1Data.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	// Verify A0 signature
	err = tss.VerifyA0Sig(
		memberID,
		dkgContext,
		req.Round1Data.A0Sig,
		tss.PublicKey(req.Round1Data.CoefficientsCommit[0]),
	)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	// Add commits to calculate accumulated commits for each index
	err = k.AddCommits(ctx, groupID, req.Round1Data.CoefficientsCommit)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrAddCommit, err.Error())
	}

	// Add Round1Data
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
		group.PubKey = tss.PublicKey(k.GetAccumulatedCommit(ctx, groupID, 0))
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

	// Compute and store its own public key
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Compute own public key
	accCommits := k.GetAllAccumulatedCommits(ctx, groupID)
	ownPubKey, err := tss.ComputeOwnPublicKey(accCommits, memberID)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrComputeOwnPubKeyFailed,
			"compute own public key failed; %s",
			err,
		)
	}

	// Update public key of the member
	member.PubKey = ownPubKey
	k.SetMember(ctx, groupID, memberID, member)

	// Set Round2Data
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

// Complain handles complaints from a member of a certain group.
// It validates the group status, verifies the member making the complaint, and checks previous submissions from the member.
// Each complaint in the request is verified, marking the member either as malicious or the subject of complaint as malicious depending on the verification result.
// After each verification, complaint statuses are appended, relevant events are emitted, and the group data is updated.
// If all members have sent confirmation or complaints, it will handle the fallen group.
func (k Keeper) Complain(
	goCtx context.Context,
	req *types.MsgComplain,
) (*types.MsgComplainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Complains[0].I

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.ROUND_3 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 3")
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
			// Mark i as malicious
			err := k.MarkMalicious(ctx, groupID, c.I)
			if err != nil {
				return nil, err
			}

			// Add complain status
			complainsWithStatus = append(complainsWithStatus, types.ComplainWithStatus{
				Complain:       c,
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
			// Mark j as malicious
			err := k.MarkMalicious(ctx, groupID, c.J)
			if err != nil {
				return nil, err
			}

			// Add complain status
			complainsWithStatus = append(complainsWithStatus, types.ComplainWithStatus{
				Complain:       c,
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
		k.SetComplainsWithStatus(ctx, groupID, types.ComplainsWithStatus{
			MemberID:            memberID,
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

// Confirm method handles a member's confirmation for a certain group.
// It validates the group status, verifies the member making the confirmation, and checks if the member has already submitted a confirmation or complaint.
// If the member's own public key signature is verified, the member's public key is updated in the data.
// If all members have sent confirmations or complaints, it will update the group status to active or handle the fallen group if there are any malicious members.
// Finally, the function sets the confirmation with status, emits an event for confirmation success, and returns a response to the confirmation request.
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
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 3")
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

	// Get confirm complain count
	confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)

	// Get malicious members
	maliciousMembers, err := k.GetMaliciousMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: req.OwnPubKeySig,
	})

	// Handle fallen group if everyone sends confirm or complains already.
	if confirmComplainCount+1 == group.Size_ {
		if len(maliciousMembers) == 0 {
			// Update group status
			group.Status = types.ACTIVE
			k.UpdateGroup(ctx, groupID, group)

			// Emit event round 3 success
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeRound3Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyStatus, group.Status.String()),
				),
			)
		} else {
			// Handle fallen group if someone in this group is malicious.
			k.handleFallenGroup(ctx, groupID, group)

			return nil, sdkerrors.Wrapf(
				types.ErrConfirmFailed,
				"memberIDs: %v is malicious",
				maliciousMembers,
			)
		}

		// Delete all dkg interim data
		k.DeleteAllDKGInterimData(ctx, groupID, group.Size_)
	}

	// Emit event confirm success
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmSuccess,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyOwnPubKeySig, hex.EncodeToString(req.OwnPubKeySig)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
		),
	)

	return &types.MsgConfirmResponse{}, nil
}

func (k Keeper) SubmitDEPairs(
	goCtx context.Context,
	req *types.MsgSubmitDEPairs,
) (*types.MsgSubmitDEPairsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accMember, err := sdk.AccAddressFromBech32(req.Member)
	if err != nil {
		return nil, err
	}

	k.HandleSetDEPairs(ctx, accMember, req.DEPairs)

	return &types.MsgSubmitDEPairsResponse{}, nil
}

// Check already confirm or complain.
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

func (k Keeper) RequestSign(goCtx context.Context, req *types.MsgRequestSign) (*types.MsgRequestSignResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get group
	group, err := k.GetGroup(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// check group status
	if group.Status != types.ACTIVE {
		return nil, sdkerrors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	members, err := k.GetMembers(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// compute bytes
	var bytes []byte
	for i, m := range members {
		accMember, err := sdk.AccAddressFromBech32(m.Member)
		if err != nil {
			return nil, err
		}
		deQueue := k.GetDEQueue(ctx, accMember)
		de, err := k.GetDE(ctx, accMember, deQueue.Head)
		if err != nil {
			return nil, err
		}

		bytes = append(bytes, sdk.Uint64ToBigEndian(uint64(i+1))...)
		bytes = append(bytes, de.PubD...)
		bytes = append(bytes, de.PubE...)
	}

	var los tss.Scalars
	var ownPubNonces tss.PublicKeys
	for i, m := range members {
		accMember, err := sdk.AccAddressFromBech32(m.Member)
		if err != nil {
			return nil, err
		}
		de, err := k.PollDEPairs(ctx, accMember)
		if err != nil {
			return nil, err
		}

		// compute own lo
		lo := tss.ComputeOwnLo(tss.MemberID(i+1), req.Message, bytes)
		los = append(los, lo)

		// compute own public nonce
		opn, err := tss.ComputeOwnPublicNonce(de.PubD, de.PubE, lo)
		if err != nil {
			return nil, err
		}
		ownPubNonces = append(ownPubNonces, opn)
	}

	groupPubNonce, err := tss.ComputeGroupPublicNonce(ownPubNonces...)
	if err != nil {
		return nil, err
	}

	// random assigning participants
	assignedParticipants, err := k.GetRandomAssigningParticipants(
		ctx,
		k.GetSigningCount(ctx)+1,
		group.Size_,
		group.Threshold,
	)
	if err != nil {
		return nil, err
	}

	signing := types.Signing{
		GroupID:              req.GroupID,
		AssignedParticipants: assignedParticipants,
		Message:              req.Message,
		GroupPubNonce:        groupPubNonce,
		Bytes:                bytes,
		Los:                  los,
		OwnPubNonces:         ownPubNonces,
		Sig:                  nil,
	}

	// set signing
	signingID := k.SetSigning(ctx, signing)

	for _, p := range assignedParticipants {
		accMember, err := sdk.AccAddressFromBech32(members[p-1].Member)
		if err != nil {
			return nil, err
		}
		k.SetPendingSign(ctx, accMember, signingID)
	}

	// emit request sign event
	event := sdk.NewEvent(
		types.EventTypeRequestSign,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", req.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
		sdk.NewAttribute(types.AttributeKeyAssignedParticipants, fmt.Sprintf("%v", assignedParticipants)),
		sdk.NewAttribute(types.AttributeBytes, hex.EncodeToString(bytes)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(groupPubNonce)),
	)
	for _, opn := range ownPubNonces {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyOwnPubNonces, hex.EncodeToString(opn)))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgRequestSignResponse{}, nil
}

func (k Keeper) Sign(goCtx context.Context, req *types.MsgSign) (*types.MsgSignResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signing, err := k.GetSigning(ctx, req.SigningID)
	if err != nil {
		return nil, err
	}

	// check signing already have signature
	if signing.Sig != nil {
		return nil, fmt.Errorf("signing ID: %d is already have signature", req.SigningID)
	}

	group, err := k.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return nil, err
	}

	member, err := k.GetMember(ctx, signing.GroupID, req.MemberID)
	if err != nil {
		return nil, err
	}

	// check sender not in assigned participants
	var found bool
	for _, ap := range signing.AssignedParticipants {
		if ap == req.MemberID {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("member ID: %d is not in assigned participants", req.MemberID)
	}

	lagrange := tss.ComputeLagrangeCoefficient(req.MemberID, signing.AssignedParticipants)

	// proof z_i
	err = tss.VerifySigningSig(
		signing.GroupPubNonce,
		group.PubKey,
		signing.Message,
		lagrange,
		req.Zi,
		member.PubKey,
	)
	if err != nil {
		return nil, err
	}

	k.SetPartialZ(ctx, req.SigningID, req.MemberID, req.Zi)

	zCount := k.GetZCount(ctx, req.SigningID)
	if zCount == group.Threshold {
		pzs := k.GetPartialZs(ctx, req.SigningID)

		sig, err := tss.CombineSignatures(pzs...)
		if err != nil {
			return nil, err
		}

		err = tss.VerifyGroupSigningSig(group.PubKey, signing.Message, sig)
		if err != nil {
			return nil, err
		}

		// emit sign success event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventSignSuccess,
				sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
				sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(sig)),
			),
		)
	}

	accMember, err := sdk.AccAddressFromBech32(member.Member)
	if err != nil {
		return nil, err
	}
	k.DeletePendingSign(ctx, accMember, req.SigningID)

	// emit submit sign event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventSubmitSign,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", req.MemberID)),
			sdk.NewAttribute(types.AttributeKeyZi, hex.EncodeToString(req.Zi)),
		),
	)

	return &types.MsgSignResponse{}, nil
}
