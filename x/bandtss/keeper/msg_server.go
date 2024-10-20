package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

type msgServer struct {
	*Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// TransitionGroup initializes a new group and setup new transition process.
func (k msgServer) TransitionGroup(
	goCtx context.Context,
	req *types.MsgTransitionGroup,
) (*types.MsgTransitionGroupResponse, error) {
	if k.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", k.authority, req.Authority)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate members
	var members []sdk.AccAddress
	for _, m := range req.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid member address: %s", err)
		}
		members = append(members, address)
	}

	// validate transition duration
	if err := k.ValidateTransitionExecTime(ctx, req.ExecTime); err != nil {
		return nil, err
	}

	// validate if transition is in progress
	if err := k.ValidateTransitionInProgress(ctx); err != nil {
		return nil, err
	}

	groupID, err := k.tssKeeper.CreateGroup(
		ctx,
		members,
		req.Threshold,
		types.ModuleName,
	)
	if err != nil {
		return nil, err
	}

	// set new group transition
	transition, err := k.SetNewGroupTransition(ctx, groupID, req.ExecTime, false)
	if err != nil {
		return nil, err
	}

	// emit an event for the group transition.
	attrs := k.ExtractEventAttributesFromTransition(transition)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGroupTransition, attrs...))

	return &types.MsgTransitionGroupResponse{}, nil
}

// ForceTransitionGroup handles the group transition without requesting a current group to
// sign a transition message.
func (k msgServer) ForceTransitionGroup(
	goCtx context.Context,
	req *types.MsgForceTransitionGroup,
) (*types.MsgForceTransitionGroupResponse, error) {
	if k.authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", k.authority, req.Authority)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate transition duration
	if err := k.ValidateTransitionExecTime(ctx, req.ExecTime); err != nil {
		return nil, err
	}

	// validate if transition is in progress
	if err := k.ValidateTransitionInProgress(ctx); err != nil {
		return nil, err
	}

	// validate incoming group
	currentGroupID := k.GetCurrentGroupID(ctx)
	if currentGroupID == req.IncomingGroupID {
		return nil, types.ErrInvalidIncomingGroup.Wrap("incoming group is the same as the current group")
	}

	incomingGroup, err := k.tssKeeper.GetGroup(ctx, req.IncomingGroupID)
	if err != nil {
		return nil, err
	}
	if incomingGroup.Status != tsstypes.GROUP_STATUS_ACTIVE {
		return nil, types.ErrInvalidIncomingGroup.Wrap("incoming group is not active")
	}

	// add members from new group.
	if err := k.AddMembers(ctx, req.IncomingGroupID); err != nil {
		return nil, err
	}

	// set new group transition
	transition, err := k.SetNewGroupTransition(ctx, req.IncomingGroupID, req.ExecTime, true)
	if err != nil {
		return nil, err
	}

	// emit an event for the group transition.
	attrs := k.ExtractEventAttributesFromTransition(transition)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGroupTransition, attrs...))

	return &types.MsgForceTransitionGroupResponse{}, nil
}

// RequestSignature initiates the signing process by requesting signatures from assigned members.
// It assigns members randomly, computes necessary values, and emits appropriate events.
func (k msgServer) RequestSignature(
	goCtx context.Context,
	req *types.MsgRequestSignature,
) (*types.MsgRequestSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	feePayer, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	content := req.GetContent()
	if content.IsInternal() {
		return nil, types.ErrContentNotAllowed.Wrapf(
			"order route: %s, type: %s", content.OrderRoute(), content.OrderType(),
		)
	}

	// Execute the handler to process the request.
	_, err = k.CreateDirectSigningRequest(ctx, content, req.Memo, feePayer, req.FeeLimit)
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestSignatureResponse{}, nil
}

// Activate update the user status back to be active
func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivate) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	if err = k.ActivateMember(ctx, sender, msg.GroupID); err != nil {
		return nil, err
	}

	return &types.MsgActivateResponse{}, nil
}

// Heartbeat keeps notice that user is alive.
func (k msgServer) Heartbeat(
	goCtx context.Context,
	msg *types.MsgHeartbeat,
) (*types.MsgHeartbeatResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	if err = k.SetLastActive(ctx, sender, msg.GroupID); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeHeartbeat,
		sdk.NewAttribute(types.AttributeKeyAddress, msg.Sender),
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", msg.GroupID)),
	))

	return &types.MsgHeartbeatResponse{}, nil
}

// UpdateParams update the parameter of the module.
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
