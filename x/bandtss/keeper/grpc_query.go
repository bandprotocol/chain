package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// IsGrantee queries if a specific address is a grantee of another.
func (q queryServer) IsGrantee(
	goCtx context.Context,
	req *types.QueryIsGranteeRequest,
) (*types.QueryIsGranteeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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

// Member queries the member information of a given account address.
func (q queryServer) Member(
	goCtx context.Context,
	req *types.QueryMemberRequest,
) (*types.QueryMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid grantee address: %s", err)
	}

	member, err := q.k.GetMember(ctx, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryMemberResponse{
		Member: member,
	}, nil
}

// Members queries filtered members information based on criteria.
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
			switch req.Status {
			case types.MEMBER_STATUS_FILTER_UNSPECIFIED:
				return m, nil
			case types.MEMBER_STATUS_FILTER_ACTIVE:
				if m.IsActive {
					return m, nil
				}
			case types.MEMBER_STATUS_FILTER_INACTIVE:
				if !m.IsActive {
					return m, nil
				}
			}

			return nil, nil
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

// CurrentGroup queries the current group information.
func (q queryServer) CurrentGroup(
	goCtx context.Context,
	req *types.QueryCurrentGroupRequest,
) (*types.QueryCurrentGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupID := q.k.GetCurrentGroupID(ctx)
	if groupID == 0 {
		return nil, types.ErrNoActiveGroup
	}

	group, err := q.k.tssKeeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &types.QueryCurrentGroupResponse{
		GroupID:   groupID,
		Size_:     group.Size_,
		Threshold: group.Threshold,
		PubKey:    group.PubKey,
		Status:    group.Status,
	}, nil
}

// Replacement queries group replacement information.
func (q queryServer) Replacement(
	goCtx context.Context,
	req *types.QueryReplacementRequest,
) (*types.QueryReplacementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	replacement := q.k.GetReplacement(ctx)

	return &types.QueryReplacementResponse{
		Replacement: replacement,
	}, nil
}

// Signing queries the bandtss signing information.
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

// Params queries parameters of bandtss module
func (q queryServer) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
