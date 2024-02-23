package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ k *Keeper }

func NewQueryServer(k *Keeper) types.QueryServer {
	return queryServer{k: k}
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
		return nil, types.ErrInvalidAccAddressFormat.Wrapf("invalid account address: %s", err)
	}

	// Get status of the address
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
			// filter item out if the request status is valid and it is not equal to the request status.
			if types.ValidMemberStatus(req.Status) && s.Status != req.Status {
				return nil, nil
			}
			return s, nil
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
