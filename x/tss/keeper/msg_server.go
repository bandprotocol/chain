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

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateGroup initializes a new group. It validates the group size, creates a new group,
// sets group members, hashes groupID and LastCommitHash to form the DKGContext, and emits
// an event for group creation.
func (k msgServer) CreateGroup(
	goCtx context.Context,
	req *types.MsgCreateGroup,
) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate group size
	groupSize := uint64(len(req.Members))
	maxGroupSize := k.MaxGroupSize(ctx)
	if groupSize > maxGroupSize {
		return nil, sdkerrors.Wrap(types.ErrGroupSizeTooLarge, fmt.Sprintf("group size exceeds %d", maxGroupSize))
	}

	// Create new group
	fee := req.Fee.Sort()
	groupID := k.CreateNewGroup(ctx, types.Group{
		Size_:     groupSize,
		Threshold: req.Threshold,
		PubKey:    nil,
		Fee:       fee,
		Status:    types.GROUP_STATUS_ROUND_1,
	})

	// Set members
	for i, m := range req.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			return nil, sdkerrors.Wrapf(
				types.ErrInvalidAccAddressFormat,
				"invalid account address: %s", err,
			)
		}

		status := k.GetStatus(ctx, address)
		if status.Status != types.MEMBER_STATUS_ACTIVE {
			return nil, types.ErrStatusIsNotActive
		}

		// ID start from 1
		k.SetMember(ctx, groupID, types.Member{
			MemberID:    tss.MemberID(i + 1),
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

// SubmitDKGRound1 validates the group status, member, coefficients commit length, one-time
// signature, and A0 signature for a group's round 1. If all checks pass, it updates the
// accumulated commits, stores the Round1Info, emits an event, and if necessary, updates the
// group status to round 2.
func (k msgServer) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round1Info.MemberID

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Check round status
	if group.Status != types.GROUP_STATUS_ROUND_1 {
		return nil, sdkerrors.Wrap(types.ErrInvalidStatus, "group status is not round 1")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member
	if !member.Verify(req.Member) {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Member,
		)
	}

	// Check previous submit
	_, err = k.GetRound1Info(ctx, groupID, req.Round1Info.MemberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 1")
	}

	// Check coefficients commit length
	if uint64(len(req.Round1Info.CoefficientCommits)) != group.Threshold {
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
	err = tss.VerifyOneTimeSig(memberID, dkgContext, req.Round1Info.OneTimeSig, req.Round1Info.OneTimePubKey)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyOneTimeSigFailed, err.Error())
	}

	// Verify A0 signature
	err = tss.VerifyA0Sig(
		memberID,
		dkgContext,
		req.Round1Info.A0Sig,
		req.Round1Info.CoefficientCommits[0],
	)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrVerifyA0SigFailed, err.Error())
	}

	// Add commits to calculate accumulated commits for each index
	err = k.AddCommits(ctx, groupID, req.Round1Info.CoefficientCommits)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrAddCommit, err.Error())
	}

	// Add round 1 info
	k.AddRound1Info(ctx, groupID, req.Round1Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			sdk.NewAttribute(
				types.AttributeKeyRound1Info,
				hex.EncodeToString(k.cdc.MustMarshal(&req.Round1Info)),
			),
		),
	)

	count := k.GetRound1InfoCount(ctx, groupID)
	if count == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroups(ctx, groupID)
	}

	return &types.MsgSubmitDKGRound1Response{}, nil
}

// SubmitDKGRound2 checks the group status, member, and whether the member has already submitted round 2 info.
// It verifies the member, checks the length of encrypted secret shares, computes and stores the member's own public key,
// sets the round 2 info, and emits appropriate events. If all members have submitted round 2 info,
// it updates the group status to round 3.
func (k msgServer) SubmitDKGRound2(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound2,
) (*types.MsgSubmitDKGRound2Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round2Info.MemberID

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Check round status
	if group.Status != types.GROUP_STATUS_ROUND_2 {
		return nil, sdkerrors.Wrap(types.ErrInvalidStatus, "group status is not round 2")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member
	if !member.Verify(req.Member) {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Member,
		)
	}

	// Check previous submit
	_, err = k.GetRound2Info(ctx, groupID, memberID)
	if err == nil {
		return nil, sdkerrors.Wrap(types.ErrAlreadySubmit, "this member already submit round 2")
	}

	// Check encrypted secret shares length
	if uint64(len(req.Round2Info.EncryptedSecretShares)) != group.Size_-1 {
		return nil, sdkerrors.Wrap(
			types.ErrEncryptedSecretSharesNotCorrectLength,
			"number of encrypted secret shares is not correct",
		)
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
	k.SetMember(ctx, groupID, member)

	// Add round 2 info
	k.AddRound2Info(ctx, groupID, req.Round2Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound2,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyMember, req.Member),
			sdk.NewAttribute(types.AttributeKeyRound2Info, hex.EncodeToString(k.cdc.MustMarshal(&req.Round2Info))),
		),
	)

	count := k.GetRound2InfoCount(ctx, groupID)
	if count == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroups(ctx, groupID)
	}

	return &types.MsgSubmitDKGRound2Response{}, nil
}

// Complain checks the group status, member, and whether the member has already confirmed or complained.
// It then verifies complaints, marks malicious members, updates the group's status if necessary,
// and finally emits appropriate events.
func (k msgServer) Complain(goCtx context.Context, req *types.MsgComplain) (*types.MsgComplainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Complaints[0].Complainant

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Check round status
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, sdkerrors.Wrap(types.ErrInvalidStatus, "group status is not round 3")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member
	if !member.Verify(req.Member) {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Member,
		)
	}

	// Check already confirm or complain
	err = k.checkConfirmOrComplain(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify complaint
	var complaintsWithStatus []types.ComplaintWithStatus
	for _, c := range req.Complaints {
		err := k.HandleVerifyComplaint(ctx, groupID, c)
		if err != nil {
			// Mark complainant as malicious
			err := k.MarkMalicious(ctx, groupID, c.Complainant)
			if err != nil {
				return nil, err
			}

			// Add complaint status
			complaintsWithStatus = append(complaintsWithStatus, types.ComplaintWithStatus{
				Complaint:       c,
				ComplaintStatus: types.COMPLAINT_STATUS_FAILED,
			})

			// Emit complain failed event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainFailed,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyComplainantID, fmt.Sprintf("%d", c.Complainant)),
					sdk.NewAttribute(types.AttributeKeyRespondentID, fmt.Sprintf("%d", c.Respondent)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		} else {
			// Mark respondent as malicious
			err := k.MarkMalicious(ctx, groupID, c.Respondent)
			if err != nil {
				return nil, err
			}

			// Add complaint status
			complaintsWithStatus = append(complaintsWithStatus, types.ComplaintWithStatus{
				Complaint:       c,
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			})

			// Emit complain success event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeComplainSuccess,
					sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
					sdk.NewAttribute(types.AttributeKeyComplainantID, fmt.Sprintf("%d", c.Complainant)),
					sdk.NewAttribute(types.AttributeKeyRespondentID, fmt.Sprintf("%d", c.Respondent)),
					sdk.NewAttribute(types.AttributeKeyKeySym, hex.EncodeToString(c.KeySym)),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(c.Signature)),
					sdk.NewAttribute(types.AttributeKeyMember, req.Member),
				),
			)
		}
	}

	// Add complain with status
	k.AddComplaintsWithStatus(ctx, groupID, types.ComplaintsWithStatus{
		MemberID:             memberID,
		ComplaintsWithStatus: complaintsWithStatus,
	})

	// Get confirm complain count
	confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)

	// Handle fallen group if everyone sends confirm or complain already
	if confirmComplainCount == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroups(ctx, groupID)
	}

	return &types.MsgComplainResponse{}, nil
}

// Confirm checks the group status and verifies the member. It then verifies the member's public key signature,
// checks the count of confirmed and complained, and handles any malicious members. If all members have
// confirmed or complained, it updates the group's status if necessary, deletes all interim data, and emits
// appropriate events.
func (k msgServer) Confirm(
	goCtx context.Context,
	req *types.MsgConfirm,
) (*types.MsgConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.MemberID

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Check round status
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, sdkerrors.Wrap(types.ErrInvalidStatus, "group status is not round 3")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify member
	if !member.Verify(req.Member) {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
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

	// Add confirm
	k.AddConfirm(ctx, groupID, types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: req.OwnPubKeySig,
	})

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

	// Handle fallen group if everyone sends confirm or complain already
	if confirmComplainCount+1 == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroups(ctx, groupID)
	}

	return &types.MsgConfirmResponse{}, nil
}

// SubmitDEs receives a member's request containing Distributed Key Generation (DKG) shares (DEs).
// It converts the member's address from Bech32 to AccAddress format and then delegates the task of setting the DEs to the HandleSetDEs function.
func (k msgServer) SubmitDEs(goCtx context.Context, req *types.MsgSubmitDEs) (*types.MsgSubmitDEsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	member, err := sdk.AccAddressFromBech32(req.Member)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidAccAddressFormat,
			"invalid account address: %s", err,
		)
	}

	err = k.HandleSetDEs(ctx, member, req.DEs)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitDEsResponse{}, nil
}

// RequestSign initiates the signing process by requesting signatures from assigned members.
// It assigns participants randomly, computes necessary values, and emits appropriate events.
func (k msgServer) RequestSignature(
	goCtx context.Context,
	req *types.MsgRequestSignature,
) (*types.MsgRequestSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feePayer, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	// Handle request sign
	_, err = k.HandleRequestSign(ctx, req.GroupID, req.Message, feePayer, req.FeeLimit)
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestSignatureResponse{}, nil
}

// SubmitSignature verifies that the member and signing process are valid, and that the member hasn't already signed.
// It checks the correctness of the signature and if the threshold is met, it combines all partial signatures into a group signature.
// It then updates the signing record, deletes all interim data, and emits appropriate events.
func (k msgServer) SubmitSignature(
	goCtx context.Context,
	req *types.MsgSubmitSignature,
) (*types.MsgSubmitSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get signing
	signing, err := k.GetSigning(ctx, req.SigningID)
	if err != nil {
		return nil, err
	}

	// Get member
	member, err := k.GetMember(ctx, signing.GroupID, req.MemberID)
	if err != nil {
		return nil, err
	}

	// Verify member
	if !member.Verify(req.Member) {
		return nil, sdkerrors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			req.MemberID,
			req.Member,
		)
	}

	// Check signing is still waiting for signature
	if signing.Status != types.SIGNING_STATUS_WAITING {
		return nil, sdkerrors.Wrapf(
			types.ErrSigningAlreadySuccess, "signing ID: %d is not in waiting state", req.SigningID,
		)
	}

	// Check member is already signed
	_, err = k.GetPartialSig(ctx, req.SigningID, req.MemberID)
	if err == nil {
		return nil, sdkerrors.Wrapf(
			types.ErrAlreadySigned,
			"member ID: %d is already signed on signing ID: %d",
			req.MemberID,
			req.SigningID,
		)
	}

	// Get group
	group, err := k.GetGroup(ctx, signing.GroupID)
	if err != nil {
		return nil, err
	}

	var found bool
	var mids []tss.MemberID
	var assignedMember types.AssignedMember
	// Check sender not in assigned participants and verify signature R
	for _, am := range signing.AssignedMembers {
		mids = append(mids, am.MemberID)
		if am.MemberID == req.MemberID {
			// Found member in assigned members
			found = true
			assignedMember = am

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

	// Compute lagrange coefficient
	lagrange := tss.ComputeLagrangeCoefficient(req.MemberID, mids)

	// Verify signing signature
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

	// Set partial signature
	k.SetPartialSig(ctx, req.SigningID, req.MemberID, req.Signature)

	sigCount := k.GetSigCount(ctx, req.SigningID)
	if sigCount == group.Threshold {
		k.AddPendingProcessSignings(ctx, req.SigningID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitSignature,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", req.MemberID)),
			sdk.NewAttribute(types.AttributeKeyMember, member.Address),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(assignedMember.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(assignedMember.PubE)),
			sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(req.Signature)),
		),
	)

	return &types.MsgSubmitSignatureResponse{}, nil
}

func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivate) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	err = k.SetActive(ctx, address)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeActivate,
		sdk.NewAttribute(types.AttributeKeyMember, msg.Address),
	))

	return &types.MsgActivateResponse{}, nil
}

// checkConfirmOrComplain checks whether a specific member has already sent a "Confirm" or "Complaint" message in a given group.
// If either a confirm or a complain message from the member is found, an error is returned.
func (k msgServer) checkConfirmOrComplain(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	_, err := k.GetConfirm(ctx, groupID, memberID)
	if err == nil {
		return sdkerrors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send confirm message",
			memberID,
		)
	}
	_, err = k.GetComplaintsWithStatus(ctx, groupID, memberID)
	if err == nil {
		return sdkerrors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send complain message",
			memberID,
		)
	}
	return nil
}
