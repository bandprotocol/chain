package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/restake/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) Keys(
	c context.Context,
	req *types.QueryKeysRequest,
) (*types.QueryKeysResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	keyStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyStoreKeyPrefix)

	filteredKeys, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		keyStore,
		req.Pagination,
		func(key []byte, v *types.Key) (*types.Key, error) {
			return v, nil
		}, func() *types.Key {
			return &types.Key{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryKeysResponse{Keys: filteredKeys, Pagination: pageRes}, nil
}

func (k Querier) Rewards(
	c context.Context,
	req *types.QueryRewardsRequest,
) (*types.QueryRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var rewards []types.Reward

	stakes := k.GetStakes(ctx, address)
	for _, stake := range stakes {
		key, err := k.GetKey(ctx, stake.Key)
		if err != nil {
			return nil, err
		}
		key = k.updateRewardPerShares(ctx, key)
		stake = k.updateRewardLefts(ctx, key, stake)

		rewards = append(rewards, types.Reward{
			Key:     key.Name,
			Rewards: stake.RewardLefts,
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

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var locks []types.Lock

	stakes := k.GetActiveStakes(ctx, address)
	for _, stake := range stakes {
		locks = append(locks, types.Lock{
			Key:    stake.Key,
			Amount: stake.Amount,
		})
	}

	return &types.QueryLocksResponse{Locks: locks}, nil
}

func (k Querier) Remainder(
	c context.Context,
	req *types.QueryRemainderRequest,
) (*types.QueryRemainderResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryRemainderResponse{Remainder: k.GetRemainder(ctx)}, nil
}
