package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
	return queryServer{k: k}
}

// Counts queries the number of data sources, oracle scripts, and requests.
func (q queryServer) Counts(c context.Context, req *types.QueryCountsRequest) (*types.QueryCountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryCountsResponse{
		GroupCount:   q.k.GetGroupCount(ctx),
		SigningCount: q.k.GetSigningCount(ctx),
	}, nil
}

// Group queries information about a group.
func (q queryServer) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := tss.GroupID(req.GroupId)

	groupResult, err := q.k.GetGroupResponse(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &types.QueryGroupResponse{
		GroupResult: *groupResult,
	}, nil
}

// Groups queries groups information.
func (q queryServer) Groups(goCtx context.Context, req *types.QueryGroupsRequest) (*types.QueryGroupsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.k.storeKey)
	groupStore := prefix.NewStore(store, types.GroupStoreKeyPrefix)

	filteredGroups, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		groupStore,
		req.Pagination,
		func(key []byte, g *types.Group) (*types.GroupResult, error) {
			return q.k.GetGroupResponse(ctx, g.ID)
		}, func() *types.Group {
			return &types.Group{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGroupsResponse{
		Groups:     filteredGroups,
		Pagination: pageRes,
	}, nil
}

// Members queries members of a group.
func (q queryServer) Members(
	goCtx context.Context,
	req *types.QueryMembersRequest,
) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get members using groupID
	members, err := q.k.GetGroupMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	return &types.QueryMembersResponse{
		Members: members,
	}, nil
}

// IsGrantee queries if a specific address is a grantee of another.
func (q queryServer) IsGrantee(
	goCtx context.Context,
	req *types.QueryIsGranteeRequest,
) (*types.QueryIsGranteeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert granter and grantee addresses from Bech32 to AccAddress
	granter, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid granter address: %s", err)
	}

	grantee, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid grantee address: %s", err)
	}

	return &types.QueryIsGranteeResponse{
		IsGrantee: q.k.CheckIsGrantee(ctx, granter, grantee),
	}, nil
}

// DE queries DEs of a given address.
func (q queryServer) DE(goCtx context.Context, req *types.QueryDERequest) (*types.QueryDEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	accAddress, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	var des []types.DE
	deStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.DEsStoreKey(accAddress))
	pageRes, err := query.Paginate(deStore, req.Pagination, func(key []byte, value []byte) error {
		var de types.DE
		if err := q.k.cdc.Unmarshal(value, &de); err != nil {
			return err
		}
		des = append(des, de)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDEResponse{
		DEs:        des,
		Pagination: pageRes,
	}, nil
}

// PendingGroups queries pending groups creation that waits a given address to submit a message.
func (q queryServer) PendingGroups(
	goCtx context.Context,
	req *types.QueryPendingGroupsRequest,
) (*types.QueryPendingGroupsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the ID of the last expired group
	lastExpired := q.k.GetLastExpiredGroupID(ctx)

	// Get the total group count
	groupCount := q.k.GetGroupCount(ctx)

	var pendingGroups []uint64
	for gid := lastExpired + 1; uint64(gid) <= groupCount; gid++ {
		// Retrieve the group object
		group := q.k.MustGetGroup(ctx, gid)

		// Check if address is the member
		member, err := q.k.GetMemberByAddress(ctx, gid, req.Address)
		if err != nil {
			continue
		}

		isSubmitted := true

		// Check submit for round 1
		if group.Status == types.GROUP_STATUS_ROUND_1 {
			if _, err := q.k.GetRound1Info(ctx, gid, member.ID); err != nil {
				isSubmitted = false
			}
		}

		// Check submit for round 2
		if group.Status == types.GROUP_STATUS_ROUND_2 {
			if _, err := q.k.GetRound2Info(ctx, gid, member.ID); err != nil {
				isSubmitted = false
			}
		}

		// Check submit for round 3 (confirm and complain)
		if group.Status == types.GROUP_STATUS_ROUND_3 {
			confirmed := true
			if _, err := q.k.GetConfirm(ctx, gid, member.ID); err != nil {
				confirmed = false
			}

			complained := true
			if _, err := q.k.GetComplaintsWithStatus(ctx, gid, member.ID); err != nil {
				complained = false
			}

			if !confirmed && !complained {
				isSubmitted = false
			}
		}

		if !isSubmitted {
			pendingGroups = append(pendingGroups, uint64(gid))
		}
	}

	return &types.QueryPendingGroupsResponse{
		PendingGroups: pendingGroups,
	}, nil
}

// PendingSignings queries signings that waits a given address to sign.
func (q queryServer) PendingSignings(
	goCtx context.Context,
	req *types.QueryPendingSigningsRequest,
) (*types.QueryPendingSigningsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	// Get pending signs.
	pendingSignings := q.k.GetPendingSignings(ctx, address)

	return &types.QueryPendingSigningsResponse{
		PendingSignings: pendingSignings,
	}, nil
}

// Signing queries signing of a given ID.
func (q queryServer) Signing(
	goCtx context.Context,
	req *types.QuerySigningRequest,
) (*types.QuerySigningResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	signingID := tss.SigningID(req.SigningId)

	signingResult, err := q.k.GetSigningResult(ctx, signingID)
	if err != nil {
		return nil, err
	}

	return &types.QuerySigningResponse{
		SigningResult: *signingResult,
	}, nil
}

// Signings queries all signings.
func (q queryServer) Signings(
	goCtx context.Context,
	req *types.QuerySigningsRequest,
) (*types.QuerySigningsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(q.k.storeKey)
	signingStore := prefix.NewStore(store, types.SigningStoreKeyPrefix)

	filteredSignings, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		signingStore,
		req.Pagination,
		func(key []byte, s *types.Signing) (*types.SigningResult, error) {
			return q.k.GetSigningResult(ctx, s.ID)
		}, func() *types.Signing {
			return &types.Signing{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySigningsResponse{
		SigningResults: filteredSignings,
		Pagination:     pageRes,
	}, nil
}

// Params queries all params of the module.
func (q queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
