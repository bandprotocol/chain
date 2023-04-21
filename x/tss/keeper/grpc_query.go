package keeper

import (
	"context"

	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	group := k.GetGroup(ctx, uint64(req.GroupId))
	return &types.QueryGroupResponse{
		Group: &group,
	}, nil
}

func (k Keeper) Members(goCtx context.Context, req *types.QueryMembersRequest) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	members := k.GetMembers(ctx, uint64(req.GroupId))
	return &types.QueryMembersResponse{
		Members: members,
	}, nil
}
