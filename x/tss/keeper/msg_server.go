package keeper

import (
	"bytes"
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

	// Validate group size
	groupSize := uint64(len(req.Members))
	maxGroupSize := k.MaxGroupSize(ctx)
	if groupSize > maxGroupSize {
		return nil, sdkerrors.Wrap(
			types.ErrGroupSizeTooLarge,
			fmt.Sprintf("group status should not more than %d", maxGroupSize),
		)
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	})

	// Set members
	for i, m := range req.Members {
		// id start from 1
		k.SetMember(ctx, groupID, tss.MemberID(i+1), types.Member{
			Address:     m,
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
		sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
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

	// check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status != types.GROUP_STATUS_ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 1")
	}

	// Verify member
	if !k.VerifyMember(ctx, groupID, memberID, req.Member) {
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
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 1")
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
		group.Status = types.GROUP_STATUS_ROUND_2
		group.PubKey = tss.PublicKey(k.GetAccumulatedCommit(ctx, groupID, 0))
		k.SetGroup(ctx, group)
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

	if group.Status != types.GROUP_STATUS_ROUND_2 {
		return nil, sdkerrors.Wrap(types.ErrRoundExpired, "group status is not round 2")
	}

	// Verify member
	if !k.VerifyMember(ctx, groupID, memberID, req.Member) {
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
		group.Status = types.GROUP_STATUS_ROUND_3
		k.SetGroup(ctx, group)
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

	if group.Status != types.GROUP_STATUS_ROUND_3 {
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
				ComplainStatus: types.COMPLAIN_STATUS_FAILED,
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
					sdk.NewAttribute(types.AttributeKeySig, hex.EncodeToString(c.Sig)),
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
				ComplainStatus: types.COMPLAIN_STATUS_SUCCESS,
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
					sdk.NewAttribute(types.AttributeKeySig, hex.EncodeToString(c.Sig)),
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
			k.handleFallenGroup(ctx, group)
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

	if group.Status != types.GROUP_STATUS_ROUND_3 {
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
			group.Status = types.GROUP_STATUS_ACTIVE
			k.SetGroup(ctx, group)

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
			k.handleFallenGroup(ctx, group)

			return nil, sdkerrors.Wrapf(
				types.ErrConfirmFailed,
				"memberIDs: %v is malicious",
				maliciousMembers,
			)
		}

		// Delete all dkg interim data
		k.DeleteAllDKGInterimData(ctx, groupID, group.Size_, group.Threshold)
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

func (k Keeper) SubmitDEs(
	goCtx context.Context,
	req *types.MsgSubmitDEs,
) (*types.MsgSubmitDEsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	accMember, err := sdk.AccAddressFromBech32(req.Member)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	k.HandleSetDEs(ctx, accMember, req.DEs)

	return &types.MsgSubmitDEsResponse{}, nil
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
	group types.Group,
) {
	group.Status = types.GROUP_STATUS_FALLEN

	k.SetGroup(ctx, group)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRound3Failed,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", group.GroupID)),
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
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return nil, sdkerrors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	members, err := k.GetMembers(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// random assigning participants
	mids, err := k.GetRandomAssigningParticipants(
		ctx,
		k.GetSigningCount(ctx)+1,
		group.Size_,
		group.Threshold,
	)
	if err != nil {
		return nil, err
	}

	// get public D and E for each asssigned members
	var assignedMembers []types.AssignedMember
	var pubDs, pubEs tss.PublicKeys
	for _, mid := range mids {
		member := members[mid-1]
		accMember, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
		}

		de, err := k.PollDE(ctx, accMember)
		if err != nil {
			return nil, err
		}

		pubDs = append(pubDs, de.PubD)
		pubEs = append(pubEs, de.PubE)

		assignedMembers = append(assignedMembers, types.AssignedMember{
			MemberID: mid,
			Member:   member.Address,
			PubD:     de.PubD,
			PubE:     de.PubE,
			PubNonce: nil,
		})
	}

	// compute bytes from mids, public D and public E
	var bytes []byte
	bytes, err = tss.ComputeBytes(mids, pubDs, pubEs)
	if err != nil {
		return nil, err
	}

	// compute lo and public nonce of each assigned member
	var ownPubNonces tss.PublicKeys
	for i, member := range assignedMembers {
		// compute own lo
		lo := tss.ComputeOwnLo(member.MemberID, req.Message, bytes)

		// compute own public nonce
		opn, err := tss.ComputeOwnPubNonce(member.PubD, member.PubE, lo)
		if err != nil {
			return nil, err
		}
		ownPubNonces = append(ownPubNonces, opn)
		assignedMembers[i].PubNonce = opn
	}

	// compute group public nonce for this signing
	groupPubNonce, err := tss.ComputeGroupPublicNonce(ownPubNonces...)
	if err != nil {
		return nil, err
	}

	signing := types.Signing{
		GroupID:         req.GroupID,
		Message:         req.Message,
		GroupPubNonce:   groupPubNonce,
		Bytes:           bytes,
		AssignedMembers: assignedMembers,
		Sig:             nil,
	}

	// add signing
	signingID := k.AddSigning(ctx, signing)

	for _, mid := range mids {
		accMember, err := sdk.AccAddressFromBech32(members[mid-1].Address)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
		}

		k.SetPendingSign(ctx, accMember, signingID)
	}

	// emit request sign event
	event := sdk.NewEvent(
		types.EventTypeRequestSign,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", req.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(req.Message)),
		sdk.NewAttribute(types.AttributeKeyBytes, hex.EncodeToString(bytes)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(groupPubNonce)),
	)
	for _, member := range assignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", member.MemberID)),
			sdk.NewAttribute(types.AttributeKeyMember, fmt.Sprintf("%s", member.Member)),
			sdk.NewAttribute(types.AttributeKeyOwnPubNonces, hex.EncodeToString(member.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(member.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(member.PubE)),
		)
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

	// check member is already signed
	_, err = k.GetPartialSig(ctx, req.SigningID, req.MemberID)
	if err == nil {
		return nil, sdkerrors.Wrapf(
			types.ErrAlreadySigned,
			"member ID: %d is already signed on signing ID: %d",
			req.MemberID,
			req.SigningID,
		)
	}

	// check signing already have signature
	if signing.Sig != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrSigningAlreadySuccess, "signing ID: %d is already have signature", req.SigningID,
		)
	}

	group, err := k.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return nil, err
	}

	member, err := k.GetMember(ctx, signing.GroupID, req.MemberID)
	if err != nil {
		return nil, err
	}

	// check sender not in assigned participants and verify signature R
	var found bool
	var mids []tss.MemberID
	for _, am := range signing.AssignedMembers {
		mids = append(mids, am.MemberID)
		if am.MemberID == req.MemberID {
			found = true

			// verify signature R
			if !bytes.Equal(req.Signature.R(), tss.Point(am.PubNonce)) {
				return nil, sdkerrors.Wrapf(
					types.ErrPubNonceNotEqualToSigR,
					"public nonce from member ID: %d is not equal signature r",
					req.MemberID,
				)
			}
		}
	}
	if !found {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAssigned, "member ID: %d is not in assigned participants", req.MemberID,
		)
	}

	lagrange := tss.ComputeLagrangeCoefficient(req.MemberID, mids)

	// verify signing signature
	err = tss.VerifySigningSig(
		signing.GroupPubNonce,
		group.PubKey,
		signing.Message,
		lagrange,
		req.Signature,
		member.PubKey,
	)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrVerifySigningSigFailed, err.Error())
	}

	k.SetPartialSig(ctx, req.SigningID, req.MemberID, req.Signature)

	sigCount := k.GetSigCount(ctx, req.SigningID)
	if sigCount == group.Threshold {
		pzs := k.GetPartialSigs(ctx, req.SigningID)

		sig, err := tss.CombineSignatures(pzs...)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrCombineSigsFailed, err.Error())
		}

		err = tss.VerifyGroupSigningSig(group.PubKey, signing.Message, sig)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrVerifyGroupSigningSigFailed, err.Error())
		}

		signing.Sig = sig
		k.SetSigning(ctx, req.SigningID, signing)

		// delete interims data
		for _, am := range signing.AssignedMembers {
			k.DeletePartialSig(ctx, req.SigningID, am.MemberID)
		}

		// emit sign success event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSignSuccess,
				sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
				sdk.NewAttribute(types.AttributeKeySig, hex.EncodeToString(sig)),
			),
		)
	}

	accMember, err := sdk.AccAddressFromBech32(member.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	k.DeletePendingSign(ctx, accMember, req.SigningID)

	// emit submit sign event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitSign,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", req.MemberID)),
			sdk.NewAttribute(types.AttributeKeySig, hex.EncodeToString(req.Signature)),
		),
	)

	return &types.MsgSignResponse{}, nil
}
