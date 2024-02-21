package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tssmember/types"
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

	input := tsstypes.UpdateGroupFeeInput{
		GroupID: req.GroupID,
		Fee:     req.Fee,
	}
	result, err := k.tssKeeper.UpdateGroupFee(ctx, input)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateGroupFee,
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", result.Group.ID)),
			sdk.NewAttribute(types.AttributeKeyFee, result.Group.Fee.String()),
		),
	)

	return &types.MsgUpdateGroupFeeResponse{}, nil
}

// CreateGroup initializes a new group. It passes the input to tss module.
func (k msgServer) CreateGroup(
	goCtx context.Context,
	req *types.MsgCreateGroup,
) (*types.MsgCreateGroupResponse, error) {
	if k.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate group size
	groupSize := uint64(len(req.Members))
	maxGroupSize := k.GetParams(ctx).MaxGroupSize
	if groupSize > maxGroupSize {
		return nil, types.ErrGroupSizeTooLarge.Wrap(fmt.Sprintf("group size exceeds %d", maxGroupSize))
	}

	// validate members
	for _, m := range req.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
		}

		status := k.tssKeeper.GetStatus(ctx, address)
		if status.Status != tsstypes.MEMBER_STATUS_ACTIVE {
			return nil, types.ErrStatusIsNotActive
		}
	}

	input := tsstypes.CreateGroupInput{
		Members:   req.Members,
		Threshold: req.Threshold,
		Fee:       req.Fee,
	}
	result, err := k.tssKeeper.CreateGroup(ctx, input)
	if err != nil {
		return nil, types.ErrCreateGroupTSSError.Wrapf("failed to create group: %s", err)
	}

	event := sdk.NewEvent(
		types.EventTypeCreateGroup,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", result.Group.ID)),
		sdk.NewAttribute(types.AttributeKeySize, fmt.Sprintf("%d", groupSize)),
		sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", input.Threshold)),
		sdk.NewAttribute(types.AttributeKeyFee, result.Group.Fee.String()),
		sdk.NewAttribute(types.AttributeKeyPubKey, ""),
		sdk.NewAttribute(types.AttributeKeyStatus, tsstypes.GROUP_STATUS_ROUND_1.String()),
		sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString(result.DKGContext)),
	)
	for _, m := range input.Members {
		event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyAddress, m))
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgCreateGroupResponse{}, nil
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
		return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
	}

	content := req.GetContent()
	if !k.router.HasRoute(content.OrderRoute()) {
		return nil, types.ErrNoSignatureOrderHandlerExists.Wrap(content.OrderRoute())
	}

	// Retrieve the appropriate handler for the request signature route.
	handler := k.router.GetRoute(content.OrderRoute())

	// Execute the handler to process the request.
	msg, err := handler(ctx, content)
	if err != nil {
		return nil, err
	}

	// Handle request sign
	signing, err := k.HandleRequestSign(ctx, req.GroupID, msg, feePayer, req.FeeLimit)
	if err != nil {
		return nil, err
	}

	event := sdk.NewEvent(
		types.EventTypeRequestSignature,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(msg)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(signing.GroupPubNonce)),
	)
	for _, am := range signing.AssignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", am.MemberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, am.Address),
			sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString(am.BindingFactor)),
			sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString(am.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(am.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(am.PubE)),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return &types.MsgRequestSignatureResponse{}, nil
}

// ReplaceGroup handles the replacement of a group with another group. It passes the input to tss module.
func (k msgServer) ReplaceGroup(
	goCtx context.Context,
	req *types.MsgReplaceGroup,
) (*types.MsgReplaceGroupResponse, error) {
	if k.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", k.authority, req.Authority)
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if NewGroupID and CurrentGroupID are active
	newGroup, err := k.tssKeeper.GetActiveGroup(ctx, req.NewGroupID)
	if err != nil {
		return nil, err
	}
	currentGroup, err := k.tssKeeper.GetActiveGroup(ctx, req.CurrentGroupID)
	if err != nil {
		return nil, err
	}

	// Verify whether the group is not in the pending replacement process.
	lastReplacementID := currentGroup.LatestReplacementID
	if lastReplacementID != uint64(0) {
		lastReplacement, err := k.tssKeeper.GetReplacement(ctx, lastReplacementID)
		if err != nil {
			return nil, err
		}

		if lastReplacement.Status == tsstypes.REPLACEMENT_STATUS_WAITING {
			return nil, types.ErrRequestReplacementFailed.Wrap(
				"the group is in the pending replacement process",
			)
		}
	}

	input := tsstypes.ReplaceGroupInput{
		CurrentGroup: currentGroup,
		NewGroup:     newGroup,
		ExecTime:     req.ExecTime,
		FeePayer:     authority,
	}
	result, err := k.tssKeeper.ReplaceGroup(ctx, input)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReplacement,
			sdk.NewAttribute(types.AttributeKeyReplacementID, fmt.Sprintf("%d", result.Replacement.ID)),
		),
	)

	return &types.MsgReplaceGroupResponse{}, nil
}

func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivate) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
	}

	if err = k.tssKeeper.SetActiveStatus(ctx, address); err != nil {
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
		return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
	}

	if err = k.tssKeeper.SetLastActive(ctx, address); err != nil {
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
