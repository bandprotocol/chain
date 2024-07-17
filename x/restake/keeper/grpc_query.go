package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bandprotocol/chain/v2/x/restake/types"
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

func (k Querier) Key(
	c context.Context,
	req *types.QueryKeyRequest,
) (*types.QueryKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	key, err := k.GetKey(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryKeyResponse{Key: key}, nil
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

	keyStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.StakesStoreKey(address))

	filteredRewards, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		keyStore,
		req.Pagination,
		func(key []byte, s *types.Stake) (*types.Reward, error) {
			reward := k.getReward(ctx, *s)
			return &reward, nil
		}, func() *types.Stake {
			return &types.Stake{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryRewardsResponse{Rewards: filteredRewards, Pagination: pageRes}, nil
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

	keyStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.StakesStoreKey(address))

	filteredLocks, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		keyStore,
		req.Pagination,
		func(key []byte, s *types.Stake) (*types.Lock, error) {
			if !k.IsActiveKey(ctx, s.Key) {
				return nil, nil
			}

			return &types.Lock{
				Key:    s.Key,
				Amount: s.Amount,
			}, nil
		}, func() *types.Stake {
			return &types.Stake{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLocksResponse{Locks: filteredLocks, Pagination: pageRes}, nil
}
