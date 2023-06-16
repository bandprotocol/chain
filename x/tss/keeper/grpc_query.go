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

// Group function handles the request to fetch information about a group.
func (k Querier) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := tss.GroupID(req.GroupId)

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get group members
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// If group is active, return group and member info
	if group.Status == types.GROUP_STATUS_ACTIVE {
		return &types.QueryGroupResponse{
			Group:   group,
			Members: members,
		}, nil
	}

	// If group is not active, get additional data
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get round infos, complaints, and confirms
	round1Infos := k.GetRound1Infos(ctx, groupID)
	round2Infos := k.GetRound2Infos(ctx, groupID)
	complaints := k.GetAllComplainsWithStatus(ctx, groupID)
	confirms := k.GetConfirms(ctx, groupID)

	// Return all the group information
	return &types.QueryGroupResponse{
		Group:                group,
		DKGContext:           dkgContext,
		Members:              members,
		Round1Infos:          round1Infos,
		Round2Infos:          round2Infos,
		ComplaintsWithStatus: complaints,
		Confirms:             confirms,
	}, nil
}

// Members function handles the request to get members of a group.
func (k Querier) Members(goCtx context.Context, req *types.QueryMembersRequest) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get members using groupID
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

	// Convert granter and grantee addresses from Bech32 to AccAddress
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

// DE function handles the request to get DEs of a given address.
func (k Querier) DE(goCtx context.Context, req *types.QueryDERequest) (*types.QueryDEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert address from Bech32 to AccAddress
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	// Get DEs and paginate the result
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

// PendingSigns function handles the request to get pending signs of a given address.
func (k Querier) PendingSigns(
	goCtx context.Context,
	req *types.QueryPendingSignsRequest,
) (*types.QueryPendingSignsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert address from Bech32 to AccAddress
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	// Get pending sign IDs and then fetch each pending sign
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

// Signings function handles the request to get signings of a given ID.
func (k Querier) Signings(
	goCtx context.Context,
	req *types.QuerySigningsRequest,
) (*types.QuerySigningsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	signingID := tss.SigningID(req.Id)

	// Get signing and partial sigs using signingID
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	pzs := k.GetPartialSigsWithKey(ctx, signingID)

	return &types.QuerySigningsResponse{
		Signing:             &signing,
		ReceivedPartialSigs: pzs,
	}, nil
}
