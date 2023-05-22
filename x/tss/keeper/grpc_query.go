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

func (k Querier) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	group, err := k.GetGroup(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return &types.QueryGroupResponse{}, err
	}

	members, err := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return &types.QueryGroupResponse{}, err
	}

	dkgContext, err := k.GetDKGContext(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return &types.QueryGroupResponse{}, err
	}

	allRound1Commitments := k.GetAllRound1Commitments(ctx, tss.GroupID(req.GroupId), group.Size_)

	return &types.QueryGroupResponse{
		Group:             &group,
		DKGContext:        dkgContext,
		Members:           members,
		Round1Commitments: allRound1Commitments,
	}, nil
}

func (k Querier) Members(goCtx context.Context, req *types.QueryMembersRequest) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	members, err := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return &types.QueryMembersResponse{}, err
	}

	return &types.QueryMembersResponse{
		Members: members,
	}, nil
}

func (k Querier) IsGrantee(
	goCtx context.Context,
	req *types.QueryIsGranteeRequest,
) (*types.QueryIsGranteeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	granter, err := sdk.AccAddressFromBech32(req.GranterAddress)
	if err != nil {
		return &types.QueryIsGranteeResponse{}, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}
	grantee, err := sdk.AccAddressFromBech32(req.GranteeAddress)
	if err != nil {
		return &types.QueryIsGranteeResponse{}, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	return &types.QueryIsGranteeResponse{
		IsGrantee: k.Keeper.IsGrantee(ctx, granter, grantee),
	}, nil
}
