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

	group, found := k.GetGroup(ctx, tss.GroupID(req.GroupId))
	if !found {
		return &types.QueryGroupResponse{}, sdkerrors.Wrapf(types.ErrGroupNotFound, "groupID: %d")
	}

	members, found := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if !found {
		return &types.QueryGroupResponse{}, sdkerrors.Wrapf(types.ErrMemberNotFound, "groupID: %d")
	}

	dkgContext, found := k.GetDKGContext(ctx, tss.GroupID(req.GroupId))
	if !found {
		return &types.QueryGroupResponse{}, sdkerrors.Wrapf(types.ErrDKGContextNotFound, "groupID: %d")
	}

	allRound1Commitments, found := k.GetAllRound1Commitments(ctx, tss.GroupID(req.GroupId))
	if !found {
		return &types.QueryGroupResponse{}, sdkerrors.Wrapf(types.ErrRound1CommitmentsNotFound, "groupID: %d")
	}

	return &types.QueryGroupResponse{
		Group:                &group,
		DKGContext:           dkgContext,
		Members:              members,
		AllRound1Commitments: allRound1Commitments,
	}, nil
}

func (k Querier) Members(goCtx context.Context, req *types.QueryMembersRequest) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	members, found := k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if !found {
		return &types.QueryMembersResponse{}, sdkerrors.Wrapf(types.ErrMemberNotFound, "groupID: %d")
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
