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
	ctx := sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	rewards := []types.Reward{}

	locks := k.GetLocks(ctx, address)
	for _, lock := range locks {
		key, err := k.GetKey(ctx, lock.Key)
		if err != nil {
			return nil, err
		}
		key = k.updateRewardPerShares(ctx, key)
		lock = k.updateRewardLefts(ctx, key, lock)

		rewards = append(rewards, types.Reward{
			Key:     key.Name,
			Rewards: lock.RewardLefts,
		})
	}

	return &types.QueryRewardsResponse{
		Rewards: rewards,
	}, nil
}

func (k Querier) Locks(
	c context.Context,
	req *types.QueryLocksRequest,
) (*types.QueryLocksResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	locks := k.GetLocks(ctx, address)
	return &types.QueryLocksResponse{Locks: locks}, nil
}
