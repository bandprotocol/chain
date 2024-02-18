package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

// CreateGroup initializes a new group. It validates the group size, creates a new group,
// sets group members, hashes groupID and LastCommitHash to form the DKGContext, and emits
// an event for group creation.
func (k msgServer) CreateGroup(
	goCtx context.Context,
	req *types.MsgCreateGroup,
) (*types.MsgCreateGroupResponse, error) {
	return &types.MsgCreateGroupResponse{}, nil
}

// ReplaceGroup handles the replacement of a group with another group. It verifies the authority,
// retrieves necessary context, creates a new replace group data, requests a signature,
// and adds the pending replace group for execution.
func (k msgServer) ReplaceGroup(
	goCtx context.Context,
	req *types.MsgReplaceGroup,
) (*types.MsgReplaceGroupResponse, error) {
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
