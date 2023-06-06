package keeper

import (
	"context"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Group function handles the request to fetch group details.
func (k Querier) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := tss.GroupID(req.GroupId)

	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	if group.Status == types.ACTIVE {
		return &types.QueryGroupResponse{
			Group:   group,
			Members: members,
		}, nil
	}

	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, err
	}

	r1s := k.GetAllRound1Data(ctx, groupID)
	r2s := k.GetAllRound2Data(ctx, groupID)
	complains := k.GetAllComplainsWithStatus(ctx, groupID)
	confirms := k.GetConfirms(ctx, groupID)

	return &types.QueryGroupResponse{
		Group:                  group,
		DKGContext:             dkgContext,
		Members:                members,
		AllRound1Data:          r1s,
		AllRound2Data:          r2s,
		AllComplainsWithStatus: complains,
		AllConfirm:             confirms,
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

func (k Querier) DE(goCtx context.Context, req *types.QueryDERequest) (*types.QueryDEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	var des []types.DE
	deStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.DEStoreKey(address))
	pageRes, err := query.Paginate(deStore, req.Pagination, func(key []byte, value []byte) error {
		var de types.DE
		k.cdc.MustUnmarshal(value, &de)
		des = append(des, de)
		return nil
	})
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidArgument, "paginate: %v", err)
	}

	return &types.QueryDEResponse{
		DEs:        des,
		Pagination: pageRes,
	}, nil
}

func (k Querier) PendingSigns(
	goCtx context.Context,
	req *types.QueryPendingSignsRequest,
) (*types.QueryPendingSignsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	var pendingSigns []types.Signing
	pendingSignIDs := k.GetPendingSignIDs(ctx, address)
	for _, id := range pendingSignIDs {
		signing, err := k.GetSigning(ctx, tss.SigningID(id))
		if err != nil {
			return nil, err
		}
		pendingSigns = append(pendingSigns, signing)
	}

	return &types.QueryPendingSignsResponse{
		PendingSigns: pendingSigns,
	}, nil
}

func (k Querier) Signings(
	goCtx context.Context,
	req *types.QuerySigningsRequest,
) (*types.QuerySigningsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signing, err := k.GetSigning(ctx, tss.SigningID(req.Id))
	if err != nil {
		return nil, err
	}
	return &types.QuerySigningsResponse{
		Signing: &signing,
	}, nil
}
