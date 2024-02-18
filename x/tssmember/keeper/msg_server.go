package keeper

import (
	"context"

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

// CreateGroup initializes a new group. It passes the input to tss module.
func (k msgServer) CreateGroup(
	goCtx context.Context,
	req *types.MsgCreateGroup,
) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	input := tsstypes.CreateGroupInput{
		Members:   req.Members,
		Threshold: req.Threshold,
		Fee:       req.Fee,
		Authority: req.Authority,
	}
	if _, err := k.tssKeeper.CreateGroup(ctx, input); err != nil {
		return nil, err
	}

	return &types.MsgCreateGroupResponse{}, nil
}

// ReplaceGroup handles the replacement of a group with another group. It passes the input to tss module.
func (k msgServer) ReplaceGroup(
	goCtx context.Context,
	req *types.MsgReplaceGroup,
) (*types.MsgReplaceGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	input := tsstypes.ReplaceGroupInput{
		CurrentGroupID: req.CurrentGroupID,
		NewGroupID:     req.NewGroupID,
		ExecTime:       req.ExecTime,
		Authority:      req.Authority,
	}
	if _, err := k.tssKeeper.ReplaceGroup(ctx, input); err != nil {
		return nil, err
	}
	return &types.MsgReplaceGroupResponse{}, nil
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
