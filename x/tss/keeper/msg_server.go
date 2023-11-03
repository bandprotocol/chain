package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate group size
	groupSize := uint64(len(req.Members))
	maxGroupSize := k.GetParams(ctx).MaxGroupSize
	if groupSize > maxGroupSize {
		return nil, errors.Wrap(types.ErrGroupSizeTooLarge, fmt.Sprintf("group size exceeds %d", maxGroupSize))
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
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", req.Threshold)),
		sdk.NewAttribute(types.AttributeKeyFee, fee.String()),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(dkgContext)),
	)
	for _, m := range req.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyAddress, m))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateGroupResponse{}, nil
}

// ReplaceGroup handles the replacement of a group with another group. It verifies the authority,
// retrieves necessary context, creates a new replace group data, requests a signature,
// and adds the pending replace group for execution.
func (k msgServer) ReplaceGroup(
	goCtx context.Context,
	req *types.MsgReplaceGroup,
) (*types.MsgReplaceGroupResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, errors.Wrapf(
			types.ErrInvalidAccAddressFormat,
			"invalid account address: %s", err,
		)
	}

	// Get from group
	fromGroup, err := k.GetGroup(ctx, req.FromGroupID)
	if err != nil {
		return nil, err
	}

	// Verify group status
	if fromGroup.Status != types.GROUP_STATUS_ACTIVE {
		return nil, errors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	// Get to group
	toGroup, err := k.GetGroup(ctx, req.ToGroupID)
	if err != nil {
		return nil, err
	}

	// Verify group status
	if toGroup.Status != types.GROUP_STATUS_ACTIVE {
		return nil, errors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	// Verify whether the group is not in the pending replacement process.
	lastReplacementID := toGroup.LatestReplacementID
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
		fromGroup.PubKey,
		req.ToGroupID,
		address,
	)
	if err != nil {
		return nil, err
	}

	nextID := k.GetNextReplacementCount(ctx)
	k.SetReplacement(ctx, types.Replacement{
		ID:          nextID,
		SigningID:   sid,
		FromGroupID: req.FromGroupID,
		FromPubKey:  fromGroup.PubKey,
		ToGroupID:   req.ToGroupID,
		ToPubKey:    toGroup.PubKey,
		ExecTime:    req.ExecTime,
		Status:      types.REPLACEMENT_STATUS_WAITING,
	})

	k.InsertReplacementQueue(ctx, nextID, req.ExecTime)

	// Update latest replacement ID to the group
	toGroup.LatestReplacementID = nextID
	k.SetGroup(ctx, toGroup)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(types.AttributeKeyReplacementID, fmt.Sprintf("%d", nextID)),
		),
	)

	return &types.MsgReplaceGroupResponse{}, nil
}

// UpdateGroupFee updates the fee for a specific group based on the provided request.
// It performs authorization checks, retrieves the group, updates the fee, and stores
// the updated group information.
func (k msgServer) UpdateGroupFee(
	goCtx context.Context,
	req *types.MsgUpdateGroupFee,
) (*types.MsgUpdateGroupFeeResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get group
	group, err := k.GetGroup(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// Set new group fee
	group.Fee = req.Fee.Sort()
	k.SetGroup(ctx, group)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateGroupFee,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", group.GroupID)),
			sdk.NewAttribute(types.AttributeKeyFee, group.Fee.String()),
		),
	)

	return &types.MsgUpdateGroupFeeResponse{}, nil
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
		return nil, errors.Wrap(types.ErrInvalidStatus, "group status is not round 1")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify address
	if !member.Verify(req.Address) {
		return nil, errors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check previous submit
	_, err = k.GetRound1Info(ctx, groupID, req.Round1Info.MemberID)
	if err == nil {
		return nil, errors.Wrap(types.ErrMemberAlreadySubmit, "this member already submit round 1")
	}

	// Check coefficients commit length
	if uint64(len(req.Round1Info.CoefficientCommits)) != group.Threshold {
		return nil, errors.Wrap(
			types.ErrInvalidLengthCoefCommits,
			"number of coefficients commit is invalid",
		)
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, errors.Wrap(types.ErrDKGContextNotFound, "dkg-context is not found")
	}

	// Verify one time signature
	err = tss.VerifyOneTimeSignature(
		memberID,
		dkgContext,
		req.Round1Info.OneTimeSignature,
		req.Round1Info.OneTimePubKey,
	)
	if err != nil {
		return nil, errors.Wrap(types.ErrVerifyOneTimeSignatureFailed, err.Error())
	}

	// Verify A0 signature
	err = tss.VerifyA0Signature(
		memberID,
		dkgContext,
		req.Round1Info.A0Signature,
		req.Round1Info.CoefficientCommits[0],
	)
	if err != nil {
		return nil, errors.Wrap(types.ErrVerifyA0SignatureFailed, err.Error())
	}

	// Add commits to calculate accumulated commits for each index
	err = k.AddCommits(ctx, groupID, req.Round1Info.CoefficientCommits)
	if err != nil {
		return nil, errors.Wrap(types.ErrAddCoefCommit, err.Error())
	}

	// Add round 1 info
	k.AddRound1Info(ctx, groupID, req.Round1Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(
				types.AttributeKeyRound1Info,
				hex.EncodeToString(k.cdc.MustMarshal(&req.Round1Info)),
			),
		),
	)

	count := k.GetRound1InfoCount(ctx, groupID)
	if count == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroup(ctx, groupID)
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
		return nil, errors.Wrap(types.ErrInvalidStatus, "group status is not round 2")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify address
	if !member.Verify(req.Address) {
		return nil, errors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check previous submit
	_, err = k.GetRound2Info(ctx, groupID, memberID)
	if err == nil {
		return nil, errors.Wrap(types.ErrMemberAlreadySubmit, "this member already submit round 2")
	}

	// Check encrypted secret shares length
	if uint64(len(req.Round2Info.EncryptedSecretShares)) != group.Size_-1 {
		return nil, errors.Wrap(
			types.ErrInvalidLengthEncryptedSecretShares,
			"number of encrypted secret shares is invalid",
		)
	}

	// Compute own public key
	accCommits := k.GetAllAccumulatedCommits(ctx, groupID)
	ownPubKey, err := tss.ComputeOwnPublicKey(accCommits, memberID)
	if err != nil {
		return nil, errors.Wrapf(
			types.ErrComputeOwnPubKeyFailed,
			"compute own public key failed; %s",
			err,
		)
	}

	// Update public key of the member
	member.PubKey = ownPubKey
	k.SetMember(ctx, member)

	// Add round 2 info
	k.AddRound2Info(ctx, groupID, req.Round2Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound2,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyRound2Info, hex.EncodeToString(k.cdc.MustMarshal(&req.Round2Info))),
		),
	)

	count := k.GetRound2InfoCount(ctx, groupID)
	if count == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroup(ctx, groupID)
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
		return nil, errors.Wrap(types.ErrInvalidStatus, "group status is not round 3")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify address
	if !member.Verify(req.Address) {
		return nil, errors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
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
					sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
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
					sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
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
		k.AddPendingProcessGroup(ctx, groupID)
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
		return nil, errors.Wrap(types.ErrInvalidStatus, "group status is not round 3")
	}

	// Get member
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}

	// Verify address
	if !member.Verify(req.Address) {
		return nil, errors.Wrapf(
			types.ErrMemberNotAuthorized,
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
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
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
		),
	)

	// Handle fallen group if everyone sends confirm or complain already
	if confirmComplainCount+1 == group.Size_ {
		// Add the pending process group to the list of pending process groups to be processed at the endblock.
		k.AddPendingProcessGroup(ctx, groupID)
	}

	return &types.MsgConfirmResponse{}, nil
}

// SubmitDEs receives a member's request containing Distributed Key Generation (DKG) shares (DEs).
// It converts the member's address from Bech32 to AccAddress format and then delegates the task of setting the DEs to the HandleSetDEs function.
func (k msgServer) SubmitDEs(goCtx context.Context, req *types.MsgSubmitDEs) (*types.MsgSubmitDEsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	member, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrapf(
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
// It assigns members randomly, computes necessary values, and emits appropriate events.
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
	_, err = k.HandleRequestSign(ctx, req.GroupID, req.GetContent(), feePayer, req.FeeLimit)
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

	// Check signing is still waiting for signature
	if signing.Status != types.SIGNING_STATUS_WAITING {
		return nil, errors.Wrapf(
			types.ErrSigningAlreadySuccess, "signing ID: %d is not in waiting state", req.SigningID,
		)
	}

	// Check sender not in assigned member
	am, found := signing.AssignedMembers.FindAssignedMember(req.MemberID, req.Address)
	if !found {
		return nil, errors.Wrapf(
			types.ErrMemberNotAssigned, "member ID/Address: %d is not in assigned members", req.MemberID,
		)
	}

	// Verify signature R
	if !signing.AssignedMembers.VerifySignatureR(req.MemberID, req.Signature.R()) {
		return nil, errors.Wrapf(
			types.ErrPubNonceNotEqualToSigR,
			"public nonce from member ID: %d is not equal signature r",
			req.MemberID,
		)
	}

	// Check member is already signed
	_, err = k.GetPartialSignature(ctx, req.SigningID, req.MemberID)
	if err == nil {
		return nil, errors.Wrapf(
			types.ErrAlreadySigned,
			"member ID: %d is already signed on signing ID: %d",
			req.MemberID,
			req.SigningID,
		)
	}

	// Compute lagrange coefficient
	lagrange, err := tss.ComputeLagrangeCoefficient(req.MemberID, signing.AssignedMembers.MemberIDs())
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidArgument, err.Error())
	}

	// Verify signing signature
	err = tss.VerifySigningSignature(
		signing.GroupPubNonce,
		signing.GroupPubKey,
		signing.Message,
		lagrange,
		req.Signature,
		am.PubKey,
	)
	if err != nil {
		return nil, errors.Wrapf(types.ErrVerifySigningSigFailed, err.Error())
	}

	// Add partial signature
	k.AddPartialSignature(ctx, req.SigningID, req.MemberID, req.Signature)

	sigCount := k.GetSignatureCount(ctx, req.SigningID)
	if sigCount == uint64(len(signing.AssignedMembers)) {
		k.AddPendingProcessSigning(ctx, req.SigningID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitSignature,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", req.MemberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, am.Address),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(am.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(am.PubE)),
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
		sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
	))

	return &types.MsgActivateResponse{}, nil
}

func (k msgServer) HealthCheck(
	goCtx context.Context,
	msg *types.MsgHealthCheck,
) (*types.MsgHealthCheckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	err = k.SetLastActive(ctx, address)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeHealthCheck,
		sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
	))

	return &types.MsgHealthCheckResponse{}, nil
}

func (k Keeper) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			k.authority,
			req.Authority,
		)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// checkConfirmOrComplain checks whether a specific member has already sent a "Confirm" or "Complaint" message in a given group.
// If either a confirm or a complain message from the member is found, an error is returned.
func (k msgServer) checkConfirmOrComplain(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	_, err := k.GetConfirm(ctx, groupID, memberID)
	if err == nil {
		return errors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send confirm message",
			memberID,
		)
	}
	_, err = k.GetComplaintsWithStatus(ctx, groupID, memberID)
	if err == nil {
		return errors.Wrapf(
			types.ErrMemberIsAlreadyComplainOrConfirm,
			"memberID %d already send complain message",
			memberID,
		)
	}
	return nil
}
