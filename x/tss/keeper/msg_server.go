package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_1 {
		return nil, types.ErrInvalidStatus.Wrap("group status is not round 1")
	}

	// Get member and verify if the sender is in the group
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	if !member.Verify(req.Address) {
		return nil, types.ErrMemberNotAuthorized.Wrapf(
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check previous submit
	if k.HasRound1Info(ctx, groupID, req.Round1Info.MemberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrap("this member already submit round 1")
	}

	// Check coefficients commit length
	if uint64(len(req.Round1Info.CoefficientCommits)) != group.Threshold {
		return nil, types.ErrInvalidLengthCoefCommits
	}

	// Get dkg-context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, types.ErrDKGContextNotFound.Wrap("dkg-context is not found")
	}

	// Verify one time signature
	err = tss.VerifyOneTimeSignature(
		memberID,
		dkgContext,
		req.Round1Info.OneTimeSignature,
		req.Round1Info.OneTimePubKey,
	)
	if err != nil {
		return nil, types.ErrVerifyOneTimeSignatureFailed.Wrap(err.Error())
	}

	// Verify A0 signature
	err = tss.VerifyA0Signature(
		memberID,
		dkgContext,
		req.Round1Info.A0Signature,
		req.Round1Info.CoefficientCommits[0],
	)
	if err != nil {
		return nil, types.ErrVerifyA0SignatureFailed.Wrap(err.Error())
	}

	// Add commits to calculate accumulated commits for each index
	if err = k.AddCommits(ctx, groupID, req.Round1Info.CoefficientCommits); err != nil {
		return nil, types.ErrAddCoeffCommit.Wrap(err.Error())
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

	// Add to the pending process group if members submit their information.
	count := k.GetRound1InfoCount(ctx, groupID)
	if count == group.Size_ {
		k.AddPendingProcessGroup(ctx, groupID)
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
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_2 {
		return nil, types.ErrInvalidStatus.Wrap("group status is not round 2")
	}

	// Get member and verify if the sender is in the group
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	if !member.Verify(req.Address) {
		return nil, types.ErrMemberNotAuthorized.Wrapf(
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check previous submit
	if k.HasRound2Info(ctx, groupID, memberID) {
		return nil, types.ErrMemberAlreadySubmit.Wrap("this member already submit round 2")
	}

	// Check encrypted secret shares length
	if uint64(len(req.Round2Info.EncryptedSecretShares)) != group.Size_-1 {
		return nil, types.ErrInvalidLengthEncryptedSecretShares
	}

	// Compute own public key
	accCommits := k.GetAllAccumulatedCommits(ctx, groupID)
	ownPubKey, err := tss.ComputeOwnPublicKey(accCommits, memberID)
	if err != nil {
		return nil, types.ErrComputeOwnPubKeyFailed.Wrapf("compute own public key failed; %s", err)
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

	// Add to the pending process group if members submit their information.
	count := k.GetRound2InfoCount(ctx, groupID)
	if count == group.Size_ {
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

	// Get group and check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, types.ErrInvalidStatus.Wrap("group status is not round 3")
	}

	// Get member and verify if the sender is in the group
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	if !member.Verify(req.Address) {
		return nil, types.ErrMemberNotAuthorized.Wrapf(
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check already confirm or complain
	if err := k.checkConfirmOrComplain(ctx, groupID, memberID); err != nil {
		return nil, err
	}

	// Verify complaint if fail to verify, mark complainant as malicious instead.
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

	// Add to the pending process group if everyone sends confirm or complain already
	confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)
	if confirmComplainCount == group.Size_ {
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

	// Get group and check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != types.GROUP_STATUS_ROUND_3 {
		return nil, types.ErrInvalidStatus.Wrap("group status is not round 3")
	}

	// Get member and verify if the sender is in the group
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return nil, err
	}
	if !member.Verify(req.Address) {
		return nil, types.ErrMemberNotAuthorized.Wrapf(
			"memberID %d address %s is not match in this group",
			memberID,
			req.Address,
		)
	}

	// Check already confirm or complain
	if err := k.checkConfirmOrComplain(ctx, groupID, memberID); err != nil {
		return nil, err
	}

	// Verify OwnPubKeySig
	if err := k.HandleVerifyOwnPubKeySig(ctx, groupID, memberID, req.OwnPubKeySig); err != nil {
		return nil, err
	}

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

	// Add to the pending process group if everyone sends confirm or complain already
	confirmComplainCount := k.GetConfirmComplainCount(ctx, groupID)
	if confirmComplainCount == group.Size_ {
		k.AddPendingProcessGroup(ctx, groupID)
	}

	return &types.MsgConfirmResponse{}, nil
}

// SubmitDEs receives a member's request containing Distributed Key Generation (DKG) shares (DEs).
// It converts the member's address from Bech32 to AccAddress format and then delegates the task of
// setting the DEs to the HandleSetDEs function.
func (k msgServer) SubmitDEs(goCtx context.Context, req *types.MsgSubmitDEs) (*types.MsgSubmitDEsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	member, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
	}

	err = k.HandleSetDEs(ctx, member, req.DEs)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitDEsResponse{}, nil
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
	signing, err := k.GetSigning(ctx, req.SigningID)
	if err != nil {
		return nil, err
	}
	if signing.Status != types.SIGNING_STATUS_WAITING {
		return nil, types.ErrSigningAlreadySuccess.Wrapf(
			"signing ID: %d is not in waiting state", req.SigningID,
		)
	}

	// Check if sender not in assigned member
	am, found := signing.AssignedMembers.FindAssignedMember(req.MemberID, req.Address)
	if !found {
		return nil, types.ErrMemberNotAssigned.Wrapf(
			"member ID/Address: %d is not in assigned members", req.MemberID,
		)
	}

	// Verify signature R
	if !signing.AssignedMembers.VerifySignatureR(req.MemberID, req.Signature.R()) {
		return nil, types.ErrPubNonceNotEqualToSigR.Wrapf(
			"public nonce from member ID: %d is not equal signature r",
			req.MemberID,
		)
	}

	// Check member is already signed
	if k.HasPartialSignature(ctx, req.SigningID, req.MemberID) {
		return nil, types.ErrAlreadySigned.Wrapf(
			"member ID: %d is already signed on signing ID: %d",
			req.MemberID,
			req.SigningID,
		)
	}

	// Compute lagrange coefficient
	lagrange, err := tss.ComputeLagrangeCoefficient(req.MemberID, signing.AssignedMembers.MemberIDs())
	if err != nil {
		return nil, types.ErrInvalidArgument.Wrap(err.Error())
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
		return nil, types.ErrVerifySigningSigFailed.Wrap(err.Error())
	}

	// Add partial signature
	k.AddPartialSignature(ctx, req.SigningID, req.MemberID, req.Signature)

	// Check if the threshold is met, if so, add to the pending process signing.
	sigCount := k.GetPartialSignatureCount(ctx, req.SigningID)
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

// UpdateParams update parameter of the module.
func (k Keeper) UpdateParams(
	goCtx context.Context,
	req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf(
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

// checkConfirmOrComplain checks whether a specific member has already sent a "Confirm" or
// "Complaint" message in a given group. If either a confirm or a complain message from the member
// is found, an error is returned.
func (k Keeper) checkConfirmOrComplain(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	if k.HasConfirm(ctx, groupID, memberID) {
		return types.ErrMemberIsAlreadyComplainOrConfirm.Wrapf(
			"memberID %d already send confirm message",
			memberID,
		)
	}

	if k.HasComplaintsWithStatus(ctx, groupID, memberID) {
		return types.ErrMemberIsAlreadyComplainOrConfirm.Wrapf(
			"memberID %d already send complain message",
			memberID,
		)
	}
	return nil
}