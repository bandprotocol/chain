package keeper

import (
	"context"

	"cosmossdk.io/errors"
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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

// Group function handles the request to fetch information about a group.
func (q queryServer) Group(goCtx context.Context, req *types.QueryGroupRequest) (*types.QueryGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupID := tss.GroupID(req.GroupId)

	// Get group
	group, err := q.k.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Get group members
	members, err := q.k.GetMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	var statuses []types.Status
	for _, m := range members {
		address := sdk.MustAccAddressFromBech32(m.Address)
		status := q.k.GetStatus(ctx, address)
		statuses = append(statuses, status)
	}

	// Ignore error as dkgContext can be deleted
	dkgContext, _ := q.k.GetDKGContext(ctx, groupID)

	// Get round infos, complaints, and confirms
	round1Infos := q.k.GetRound1Infos(ctx, groupID)
	round2Infos := q.k.GetRound2Infos(ctx, groupID)
	complaints := q.k.GetAllComplainsWithStatus(ctx, groupID)
	confirms := q.k.GetConfirms(ctx, groupID)

	// Return all the group information
	return &types.QueryGroupResponse{
		Group:                group,
		DKGContext:           dkgContext,
		Members:              members,
		Statuses:             statuses,
		Round1Infos:          round1Infos,
		Round2Infos:          round2Infos,
		ComplaintsWithStatus: complaints,
		Confirms:             confirms,
	}, nil
}

// Members function handles the request to get members of a group.
func (q queryServer) Members(
	goCtx context.Context,
	req *types.QueryMembersRequest,
) (*types.QueryMembersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get members using groupID
	members, err := q.k.GetMembers(ctx, tss.GroupID(req.GroupId))
	if err != nil {
		return nil, err
	}

	return &types.QueryMembersResponse{
		Members: members,
	}, nil
}

// IsGrantee function handles the request to check if a specific address is a grantee of another.
func (q queryServer) IsGrantee(
	goCtx context.Context,
	req *types.QueryIsGranteeRequest,
) (*types.QueryIsGranteeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert granter and grantee addresses from Bech32 to AccAddress
	granter, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	grantee, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	return &types.QueryIsGranteeResponse{
		IsGrantee: q.k.CheckIsGrantee(ctx, granter, grantee),
	}, nil
}

// DE function handles the request to get DEs of a given address.
func (q queryServer) DE(goCtx context.Context, req *types.QueryDERequest) (*types.QueryDEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	accAddress, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAccAddressFormat, "invalid account address: %s", err)
	}

	// Get DEs and paginate the result
	var des []types.DE
	deStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.DEStoreKey(accAddress))
	pageRes, err := query.Paginate(deStore, req.Pagination, func(key []byte, value []byte) error {
		var de types.DE
		q.k.cdc.MustUnmarshal(value, &de)
		des = append(des, de)
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidArgument, "paginate: %v", err)
	}

	return &types.QueryDEResponse{
		DEs:        des,
		Pagination: pageRes,
	}, nil
}

// PendingGroups function handles the request to get pending groups of a given address.
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
			if _, err := q.k.GetRound1Info(ctx, gid, member.MemberID); err != nil {
				isSubmitted = false
			}
		}

		// Check submit for round 2
		if group.Status == types.GROUP_STATUS_ROUND_2 {
			if _, err := q.k.GetRound2Info(ctx, gid, member.MemberID); err != nil {
				isSubmitted = false
			}
		}

		// Check submit for round 3 (confirm and complain)
		if group.Status == types.GROUP_STATUS_ROUND_3 {
			confirmed := true
			if _, err := q.k.GetConfirm(ctx, gid, member.MemberID); err != nil {
				confirmed = false
			}

			complained := true
			if _, err := q.k.GetComplaintsWithStatus(ctx, gid, member.MemberID); err != nil {
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

// PendingSignings function handles the request to get pending signs of a given address.
func (q queryServer) PendingSignings(
	goCtx context.Context,
	req *types.QueryPendingSigningsRequest,
) (*types.QueryPendingSigningsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
	}

	// Get pending signs.
	pendingSignings := q.k.GetPendingSignings(ctx, address)

	return &types.QueryPendingSigningsResponse{
		PendingSignings: pendingSignings,
	}, nil
}

// Signing function handles the request to get signing of a given ID.
func (q queryServer) Signing(
	goCtx context.Context,
	req *types.QuerySigningRequest,
) (*types.QuerySigningResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	signingID := tss.SigningID(req.SigningId)

	// Get signing and partial sigs using signingID
	signing, err := q.k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	pzs := q.k.GetPartialSignaturesWithKey(ctx, signingID)

	var evmSignature *types.EVMSignature
	if signing.Signature != nil {
		rAddress, err := signing.Signature.R().Address()
		if err != nil {
			return nil, err
		}

		evmSignature = &types.EVMSignature{
			RAddress:  rAddress,
			Signature: tmbytes.HexBytes(signing.Signature.S()),
		}
	}

	return &types.QuerySigningResponse{
		Signing:                   signing,
		EVMSignature:              evmSignature,
		ReceivedPartialSignatures: pzs,
	}, nil
}

// Status function handles the request to get the status of a given account address.
func (q queryServer) Status(
	goCtx context.Context,
	req *types.QueryStatusRequest,
) (*types.QueryStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert the address from Bech32 format to AccAddress format
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAccAddressFormat, "invalid account address: %s", err)
	}

	// Get all statuses of the address
	status := q.k.GetStatus(ctx, address)

	return &types.QueryStatusResponse{
		Status: status,
	}, nil
}

// Statuses function handles the request to get filtered statuses based on criteria.
func (q queryServer) Statuses(
	goCtx context.Context,
	req *types.QueryStatusesRequest,
) (*types.QueryStatusesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	statusStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.StatusStoreKeyPrefix)

	filteredStatuses, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		statusStore,
		req.Pagination,
		func(key []byte, s *types.Status) (*types.Status, error) {
			matchStatus := true

			// Check if status is valids
			if types.ValidMemberStatus(req.Status) {
				matchStatus = s.Status == req.Status
			}

			if matchStatus {
				return s, nil
			}

			return nil, nil
		},
		func() *types.Status {
			return &types.Status{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStatusesResponse{
		Statuses:   filteredStatuses,
		Pagination: pageRes,
	}, nil
}

// Replacement function handles the request to get replacement of a given ID.
func (q queryServer) Replacement(
	goCtx context.Context,
	req *types.QueryReplacementRequest,
) (*types.QueryReplacementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get replacement
	r, err := q.k.GetReplacement(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryReplacementResponse{
		Replacement: &r,
	}, nil
}

// Replacements function handles the request to get replacements.
func (q queryServer) Replacements(
	goCtx context.Context,
	req *types.QueryReplacementsRequest,
) (*types.QueryReplacementsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	replacementStore := prefix.NewStore(ctx.KVStore(q.k.storeKey), types.ReplacementKeyPrefix)

	// Get pending replace groups
	filteredReplacements, pageRes, err := query.GenericFilteredPaginate(
		q.k.cdc,
		replacementStore,
		req.Pagination,
		func(key []byte, rg *types.Replacement) (*types.Replacement, error) {
			matchStatus := true

			// match status (if supplied/valid)
			if types.ValidReplacementStatus(req.Status) {
				matchStatus = rg.Status == req.Status
			}

			if matchStatus {
				return rg, nil
			}

			return nil, nil
		}, func() *types.Replacement {
			return &types.Replacement{}
		})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryReplacementsResponse{Replacements: filteredReplacements, Pagination: pageRes}, nil
}
