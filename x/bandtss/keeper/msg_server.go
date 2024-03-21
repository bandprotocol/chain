package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

type msgServer struct {
	*Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
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

	// validate members
	var members []sdk.AccAddress
	for _, m := range req.Members {
		address, err := sdk.AccAddressFromBech32(m)
		if err != nil {
			return nil, err
		}
		members = append(members, address)
	}

	if _, err := k.tssKeeper.CreateGroup(ctx, members, req.Threshold, req.Fee, types.ModuleName); err != nil {
		return nil, err
	}

	return &types.MsgCreateGroupResponse{}, nil
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
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	currentGroupID := k.GetCurrentGroupID(ctx)
	if req.CurrentGroupID != currentGroupID {
		return nil, types.ErrInvalidGroupID.Wrapf("invalid currentGroupID; expect %d got %d", currentGroupID, req.CurrentGroupID)
	}

	if k.GetReplacingGroupID(ctx) != 0 {
		return nil, types.ErrReplacementInProgress
	}

	_, err = k.tssKeeper.ReplaceGroup(ctx, req.CurrentGroupID, req.NewGroupID, req.ExecTime, authority, sdk.NewCoins())
	if err != nil {
		return nil, err
	}
	k.SetReplacingGroupID(ctx, req.NewGroupID)

	return &types.MsgReplaceGroupResponse{}, nil
}

// UpdateGroupFee updates the fee for a specific group based on the provided request.
func (k msgServer) UpdateGroupFee(
	goCtx context.Context,
	req *types.MsgUpdateGroupFee,
) (*types.MsgUpdateGroupFeeResponse, error) {
	if k.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := k.tssKeeper.UpdateGroupFee(ctx, req.GroupID, req.Fee)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateGroupFeeResponse{}, nil
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
		return nil, err
	}

	// Execute the handler to process the request.
	_, err = k.HandleCreateSigning(ctx, req.GroupID, req.GetContent(), feePayer, req.FeeLimit)
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestSignatureResponse{}, nil
}

// Activate update the user status back to be active
func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivate) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	if err = k.SetActiveStatuses(ctx, []sdk.AccAddress{address}); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeActivate,
		sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
	))

	return &types.MsgActivateResponse{}, nil
}

// HealthCheck keeps notice that user is alive.
func (k msgServer) HealthCheck(
	goCtx context.Context,
	msg *types.MsgHealthCheck,
) (*types.MsgHealthCheckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	if err = k.SetLastActive(ctx, address); err != nil {
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
