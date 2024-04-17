package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
	return queryServer{k: k}
}

// Member function handles the request to get the member of a given account address.
func (q queryServer) Member(
	goCtx context.Context,
	req *types.QueryMemberRequest,
) (*types.QueryMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	// Get member of the address
	member, err := q.k.GetMember(ctx, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryMemberResponse{
		Member: member,
	}, nil
}

// Members function handles the request to get filtered members based on criteria.
func (q queryServer) Members(
	goCtx context.Context,
	req *types.QueryMembersRequest,
) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	memberStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.MemberStoreKeyPrefix)
	filteredMembers, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		memberStore,
		req.Pagination,
		func(key []byte, m *types.Member) (*types.Member, error) {
			// filter item out if the member's isActive is not equal to the request status.
			if m.IsActive != req.IsActive {
				return nil, nil
			}
			return m, nil
		},
		func() *types.Member {
			return &types.Member{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryMembersResponse{
		Members:    filteredMembers,
		Pagination: pageRes,
	}, nil
}

// CurrentGroup function handles the request to get the current group information.
// TODO: update current group response later when add election flow.
func (q queryServer) CurrentGroup(
	goCtx context.Context,
	req *types.QueryCurrentGroupRequest,
) (*types.QueryCurrentGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupID := q.k.GetCurrentGroupID(ctx)
	return &types.QueryCurrentGroupResponse{
		GroupID: uint64(groupID),
	}, nil
}

// ReplacingGroup function handles the request to get the replacing group information.
// TODO: update current group response later when add election flow.
func (q queryServer) ReplacingGroup(
	goCtx context.Context,
	req *types.QueryReplacingGroupRequest,
) (*types.QueryReplacingGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupID := q.k.GetReplacingGroupID(ctx)
	return &types.QueryReplacingGroupResponse{
		GroupID: uint64(groupID),
	}, nil
}

// Replacement function handles the request to get the group replacement information.
func (q queryServer) Replacement(
	goCtx context.Context,
	req *types.QueryReplacementRequest,
) (*types.QueryReplacementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	replacement, err := q.k.GetReplacement(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryReplacementResponse{
		Replacement: replacement,
	}, nil
}

// Signing function handles the request to get the bandtss signing information.
func (q queryServer) Signing(
	goCtx context.Context,
	req *types.QuerySigningRequest,
) (*types.QuerySigningResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get signing and partial sigs using signingID
	signing, err := q.k.GetSigning(ctx, types.SigningID(req.SigningId))
	if err != nil {
		return nil, err
	}

	currentGroupSigningResult, err := q.k.tssKeeper.GetSigningResult(ctx, signing.CurrentGroupSigningID)
	if err != nil {
		return nil, err
	}

	var replacingGroupSigningResult *tsstypes.SigningResult
	if signing.ReplacingGroupSigningID != 0 {
		replacingGroupSigningResult, err = q.k.tssKeeper.GetSigningResult(ctx, signing.ReplacingGroupSigningID)
		if err != nil {
			return nil, err
		}
	}

	return &types.QuerySigningResponse{
		Fee:                         signing.Fee,
		Requester:                   signing.Requester,
		CurrentGroupSigningResult:   currentGroupSigningResult,
		ReplacingGroupSigningResult: replacingGroupSigningResult,
	}, nil
}

// Params return parameters of bandtss module
func (q queryServer) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
