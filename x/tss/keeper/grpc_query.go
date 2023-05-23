package keeper

import (
	"context"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Group function handles the request to fetch group details.
func (k Querier) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	group, err := k.GetGroup(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	members, err := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	dkgContext, err := k.GetDKGContext(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	allRound1Data := k.GetAllRound1Datas(ctx, tss.GroupID(req.GroupId), group.Size_)

	return &types.QueryGroupResponse{
		Group:      group,
		DKGContext: dkgContext,
		Members:    members,
		Round1Data: allRound1Data,
	}, nil
}

// Members function handles the request to fetch members of a group.
func (k Querier) Members(goCtx context.Context, req *types.QueryMembersRequest) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	members, err := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	return &types.QueryMembersResponse{
		Members: members,
	}, nil
}

// IsGrantee function handles the request to check if a specific address is a grantee of another.
func (k Querier) IsGrantee(
	goCtx context.Context,
	req *types.QueryIsGranteeRequest,
) (*types.QueryIsGranteeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	granter, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}
	grantee, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	return &types.QueryIsGranteeResponse{
		IsGrantee: k.Keeper.IsGrantee(ctx, granter, grantee),
	}, nil
}
