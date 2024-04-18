package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) Rewards(
	c context.Context,
	req *types.QueryRewardsRequest,
) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &types.QueryRewardsResponse{}, nil
}

func (k Querier) LockTokens(
	c context.Context,
	req *types.QueryLockTokensRequest,
) (*types.QueryLockTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &types.QueryLockTokensResponse{}, nil
}
