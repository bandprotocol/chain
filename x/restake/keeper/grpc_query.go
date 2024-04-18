package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/restake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	_ = sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &types.QueryRewardsResponse{}, nil
}

func (k Querier) Locks(
	c context.Context,
	req *types.QueryLocksRequest,
) (*types.QueryLocksResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addresss, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	locks := k.GetLocks(ctx, addresss)
	return &types.QueryLocksResponse{Locks: locks}, nil
}
