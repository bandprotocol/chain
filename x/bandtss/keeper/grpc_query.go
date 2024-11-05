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
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
	return queryServer{k: k}
}

// Counts queries the number signing requests.
func (q queryServer) Counts(c context.Context, req *types.QueryCountsRequest) (*types.QueryCountsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryCountsResponse{
		SigningCount: q.k.GetSigningCount(ctx),
	}, nil
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
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	currentGroupID := q.k.GetCurrentGroup(ctx).GroupID
	currentGroupMember, _ := q.k.GetMember(ctx, address, currentGroupID)

	incomingGroupID := q.k.GetIncomingGroupID(ctx)
	incomingGroupMember, _ := q.k.GetMember(ctx, address, incomingGroupID)

	return &types.QueryMemberResponse{
		CurrentGroupMember:  currentGroupMember,
		IncomingGroupMember: incomingGroupMember,
	}, nil
}

// Members queries filtered members information based on criteria. If queried group
// is not activated, it will return an empty list.
func (q queryServer) Members(
	goCtx context.Context,
	req *types.QueryMembersRequest,
) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var groupID tss.GroupID
	if req.IsIncomingGroup {
		groupID = q.k.GetIncomingGroupID(ctx)
	} else {
		groupID = q.k.GetCurrentGroup(ctx).GroupID
	}

	iteratorKey := append(types.MemberStoreKeyPrefix, sdk.Uint64ToBigEndian(uint64(groupID))...)
	memberStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), iteratorKey)

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

	currentGroup := q.k.GetCurrentGroup(ctx)
	if currentGroup.GroupID == 0 {
		return nil, types.ErrNoCurrentGroup
	}

	group, err := q.k.tssKeeper.GetGroup(ctx, currentGroup.GroupID)
	if err != nil {
		return nil, err
	}

	return &types.QueryCurrentGroupResponse{
		GroupID:    currentGroup.GroupID,
		Size_:      group.Size_,
		Threshold:  group.Threshold,
		PubKey:     group.PubKey,
		Status:     group.Status,
		ActiveTime: currentGroup.ActiveTime,
	}, nil
}

// IncomingGroup queries the incoming group information.
func (q queryServer) IncomingGroup(
	goCtx context.Context,
	req *types.QueryIncomingGroupRequest,
) (*types.QueryIncomingGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupID := q.k.GetIncomingGroupID(ctx)
	if groupID == 0 {
		return nil, types.ErrNoIncomingGroup
	}

	group, err := q.k.tssKeeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &types.QueryIncomingGroupResponse{
		GroupID:   groupID,
		Size_:     group.Size_,
		Threshold: group.Threshold,
		PubKey:    group.PubKey,
		Status:    group.Status,
	}, nil
}

// GroupTransition queries group transition information.
func (q queryServer) GroupTransition(
	goCtx context.Context,
	req *types.QueryGroupTransitionRequest,
) (*types.QueryGroupTransitionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	gt, found := q.k.GetGroupTransition(ctx)
	var groupTransition *types.GroupTransition
	if found {
		groupTransition = &gt
	}

	return &types.QueryGroupTransitionResponse{
		GroupTransition: groupTransition,
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

	var currentGroupSigningResult *tsstypes.SigningResult
	var incomingGroupSigningResult *tsstypes.SigningResult

	if signing.CurrentGroupSigningID != 0 {
		currentGroupSigningResult, err = q.k.tssKeeper.GetSigningResult(ctx, signing.CurrentGroupSigningID)
		if err != nil {
			return nil, err
		}
	}

	if signing.IncomingGroupSigningID != 0 {
		incomingGroupSigningResult, err = q.k.tssKeeper.GetSigningResult(ctx, signing.IncomingGroupSigningID)
		if err != nil {
			return nil, err
		}
	}

	return &types.QuerySigningResponse{
		FeePerSigner:               signing.FeePerSigner,
		Requester:                  signing.Requester,
		CurrentGroupSigningResult:  currentGroupSigningResult,
		IncomingGroupSigningResult: incomingGroupSigningResult,
	}, nil
}

// Params queries parameters of bandtss module
func (q queryServer) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(stdCtx)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}
