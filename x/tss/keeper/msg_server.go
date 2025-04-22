package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// SubmitDKGRound1 validates the group status, member, coefficients commit length, one-time
// signature, and A0 signature for a group's round 1. If all checks pass, it updates the
// accumulated commits, stores the Round1Info, emits an event, and if necessary, add to the
// pending process group for processing group status to round 2.
func (k msgServer) SubmitDKGRound1(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound1,
) (*types.MsgSubmitDKGRound1Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round1Info.MemberID

	// Get group and check group status
	group, err := k.Keeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_1 {
		return nil, types.ErrInvalidGroupStatus.Wrapf("the status of groupID %d is not round 1", groupID)
	}

	// Validate memberID
	if err := k.Keeper.ValidateMemberID(ctx, groupID, memberID, req.Sender); err != nil {
		return nil, err
	}

	// Check previous submit
	if k.Keeper.HasRound1Info(ctx, groupID, req.Round1Info.MemberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit round 1 message",
			memberID,
			groupID,
		)
	}

	if err := k.Keeper.ValidateRound1Info(ctx, group, req.Round1Info); err != nil {
		return nil, err
	}

	// Add commits to calculate accumulated commits for each index
	if err = k.Keeper.AddCoefficientCommits(ctx, groupID, req.Round1Info.CoefficientCommits); err != nil {
		return nil, err
	}

	// Add round 1 info
	k.Keeper.AddRound1Info(ctx, groupID, req.Round1Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound1,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Sender),
			sdk.NewAttribute(
				types.AttributeKeyRound1Info,
				hex.EncodeToString(k.cdc.MustMarshal(&req.Round1Info)),
			),
		),
	)

	// Add to the pending process group if members submit their information.
	count := k.Keeper.GetRound1InfoCount(ctx, groupID)
	if count == group.Size_ {
		k.Keeper.AddPendingProcessGroup(ctx, groupID)
	}

	return &types.MsgSubmitDKGRound1Response{}, nil
}

// SubmitDKGRound2 checks the group status, member, and whether the member has already
// submitted round 2 info. It verifies the member, checks the length of encrypted secret shares,
// computes and stores the member's own public key, sets the round 2 info, and emits appropriate events.
// If all members have submitted round 2 info, add to the pending process group for processing
// group status to round 3.
func (k msgServer) SubmitDKGRound2(
	goCtx context.Context,
	req *types.MsgSubmitDKGRound2,
) (*types.MsgSubmitDKGRound2Response, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := req.GroupID
	memberID := req.Round2Info.MemberID

	// Get group and check group status
	group, err := k.Keeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_2 {
		return nil, types.ErrInvalidGroupStatus.Wrapf("the status of groupID %d is not round 2", groupID)
	}

	// Validate memberID
	if err := k.Keeper.ValidateMemberID(ctx, groupID, memberID, req.Sender); err != nil {
		return nil, err
	}

	// Check previous submit
	if k.Keeper.HasRound2Info(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit round 2 message",
			memberID,
			groupID,
		)
	}

	// Check encrypted secret shares length
	if uint64(len(req.Round2Info.EncryptedSecretShares)) != group.Size_-1 {
		return nil, types.ErrInvalidLengthEncryptedSecretShares
	}

	// Update member public key of the group.
	if err := k.Keeper.UpdateMemberPubKey(ctx, groupID, memberID); err != nil {
		return nil, err
	}

	// Add round 2 info
	k.Keeper.AddRound2Info(ctx, groupID, req.Round2Info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitDKGRound2,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", memberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Sender),
			sdk.NewAttribute(types.AttributeKeyRound2Info, hex.EncodeToString(k.cdc.MustMarshal(&req.Round2Info))),
		),
	)

	// Add to the pending process group if members submit their information.
	count := k.Keeper.GetRound2InfoCount(ctx, groupID)
	if count == group.Size_ {
		k.Keeper.AddPendingProcessGroup(ctx, groupID)
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

	// Get group and check group status
	group, err := k.Keeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, types.ErrInvalidGroupStatus.Wrapf("the status of groupID %d is not round 3", groupID)
	}

	// Validate memberID
	if err := k.Keeper.ValidateMemberID(ctx, groupID, memberID, req.Sender); err != nil {
		return nil, err
	}

	// Check already confirm or complain
	if k.Keeper.HasConfirm(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit confirm message",
			memberID,
			groupID,
		)
	}
	if k.Keeper.HasComplaintsWithStatus(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit complaint message",
			memberID,
			groupID,
		)
	}

	// Verify complaint if fail to verify, mark complainant as malicious instead.
	complaintsWithStatus, err := k.Keeper.ProcessComplaint(ctx, req.Complaints, groupID, req.Sender)
	if err != nil {
		return nil, err
	}

	// Add complain with status
	k.Keeper.AddComplaintsWithStatus(ctx, groupID, types.ComplaintsWithStatus{
		MemberID:             memberID,
		ComplaintsWithStatus: complaintsWithStatus,
	})

	// Add to the pending process group if everyone sends confirm or complain already
	confirmComplainCount := k.Keeper.GetConfirmComplainCount(ctx, groupID)
	if confirmComplainCount == group.Size_ {
		k.Keeper.AddPendingProcessGroup(ctx, groupID)
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

	// Get group and check group status
	group, err := k.Keeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, types.ErrInvalidGroupStatus.Wrapf("the status of groupID %d is not round 3", groupID)
	}

	// Validate memberID
	if err := k.Keeper.ValidateMemberID(ctx, groupID, memberID, req.Sender); err != nil {
		return nil, err
	}

	// Check already confirm or complain
	if k.Keeper.HasConfirm(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit confirm message",
			memberID,
			groupID,
		)
	}
	if k.Keeper.HasComplaintsWithStatus(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrapf(
			"memberID %d in group ID %d already submit complaint message",
			memberID,
			groupID,
		)
	}

	// Verify OwnPubKeySig
	if err := k.Keeper.VerifyOwnPubKeySignature(ctx, groupID, memberID, req.OwnPubKeySig); err != nil {
		return nil, err
	}

	// Add confirm
	k.Keeper.AddConfirm(ctx, groupID, types.NewConfirm(memberID, req.OwnPubKeySig))

	// Emit event confirm success
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfirmSuccess,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", groupID)),
			sdk.NewAttribute(types.AttributeKeyOwnPubKeySig, hex.EncodeToString(req.OwnPubKeySig)),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Sender),
		),
	)

	// Add to the pending process group if everyone sends confirm or complain already
	confirmComplainCount := k.Keeper.GetConfirmComplainCount(ctx, groupID)
	if confirmComplainCount == group.Size_ {
		k.Keeper.AddPendingProcessGroup(ctx, groupID)
	}

	return &types.MsgConfirmResponse{}, nil
}

// SubmitDEs receives a member's request containing Distributed Key Generation (DKG) shares (DEs).
// It converts the member's address from Bech32 to AccAddress format and then delegates the task of
// setting the DEs to the HandleSetDEs function.
func (k msgServer) SubmitDEs(goCtx context.Context, req *types.MsgSubmitDEs) (*types.MsgSubmitDEsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	member, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	err = k.Keeper.EnqueueDEs(ctx, member, req.DEs)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitDEsResponse{}, nil
}

// ResetDE removes all DEs of the given member in the queue.
func (k msgServer) ResetDE(goCtx context.Context, req *types.MsgResetDE) (*types.MsgResetDEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	member, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	// Reset DE
	if err := k.Keeper.ResetDE(ctx, member); err != nil {
		return nil, err
	}

	return &types.MsgResetDEResponse{}, nil
}

// SubmitSignature verifies that the member and signing process are valid, and that the member
// hasn't already signed. It checks the correctness of the signature and if the threshold is met,
// it combines all partial signatures into a group signature. It then updates the signing record,
// deletes all interim data, and emits appropriate events.
func (k msgServer) SubmitSignature(
	goCtx context.Context,
	req *types.MsgSubmitSignature,
) (*types.MsgSubmitSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get signing and check signing is still waiting for signature
	signing, err := k.Keeper.GetSigning(ctx, req.SigningID)
	if err != nil {
		return nil, err
	}
	if signing.Status != types.SIGNING_STATUS_WAITING {
		return nil, types.ErrSigningAlreadySuccess.Wrapf(
			"signing ID: %d is already success", req.SigningID,
		)
	}

	sa, err := k.Keeper.GetSigningAttempt(ctx, req.SigningID, signing.CurrentAttempt)
	if err != nil {
		return nil, err
	}
	assignedMembers := types.AssignedMembers(sa.AssignedMembers)

	// validate member address.
	am, found := assignedMembers.FindAssignedMember(req.MemberID)
	if !found || am.Address != req.Signer {
		return nil, types.ErrMemberNotAssigned.Wrapf(
			"member ID %d is not in assigned members", req.MemberID,
		)
	}

	// Check member is already signed
	if k.Keeper.HasPartialSignature(ctx, req.SigningID, sa.Attempt, req.MemberID) {
		return nil, types.ErrAlreadySigned.Wrapf(
			"member ID %d already signed on signing ID: %d",
			req.MemberID,
			req.SigningID,
		)
	}

	// Verify signature R
	if !assignedMembers.VerifySignatureR(req.MemberID, req.Signature.R()) {
		return nil, types.ErrSubmitSigningSignatureFailed.Wrapf(
			"public nonce from member ID %d is not equal signature r",
			req.MemberID,
		)
	}

	// Compute lagrange coefficient
	lagrange, err := tss.ComputeLagrangeCoefficient(req.MemberID, assignedMembers.MemberIDs())
	if err != nil {
		return nil, types.ErrSubmitSigningSignatureFailed.Wrapf(
			"failed to compute lagrange coefficient: %v", err,
		)
	}

	// Verify signing signature
	if err = tss.VerifySignature(
		signing.GroupPubNonce,
		signing.GroupPubKey,
		signing.Message,
		lagrange,
		req.Signature,
		am.PubKey,
	); err != nil {
		return nil, types.ErrSubmitSigningSignatureFailed.Wrapf(
			"failed to verify signing signature: %v", err,
		)
	}

	// Add partial signature
	k.Keeper.AddPartialSignature(ctx, req.SigningID, sa.Attempt, req.MemberID, req.Signature)

	// Check if the threshold is met, if so, add to the pending process signing.
	sigCount := k.Keeper.GetPartialSignatureCount(ctx, req.SigningID, sa.Attempt)
	if sigCount == uint64(len(assignedMembers)) {
		k.Keeper.AddPendingProcessSigning(ctx, req.SigningID)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitSignature,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", req.SigningID)),
			sdk.NewAttribute(types.AttributeKeyAttempt, fmt.Sprintf("%d", signing.CurrentAttempt)),
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

// UpdateParams update parameter of the module.
func (k msgServer) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	if k.Keeper.GetAuthority() != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
			"invalid authority; expected %s, got %s",
			k.Keeper.GetAuthority(),
			req.Authority,
		)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
