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

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

// Keys queries all keys with pagination.
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

// Key queries info about a key.
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

// Rewards queries all rewards with pagination.
func (k Querier) Rewards(
	c context.Context,
	req *types.QueryRewardsRequest,
) (*types.QueryRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.LockerAddress)
	if err != nil {
		return nil, err
	}

	lockStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.LocksByAddressStoreKey(addr))

	filteredRewards, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		lockStore,
		req.Pagination,
		func(key []byte, s *types.Lock) (*types.Reward, error) {
			reward := k.getReward(ctx, *s)
			return &reward, nil
		}, func() *types.Lock {
			return &types.Lock{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryRewardsResponse{Rewards: filteredRewards, Pagination: pageRes}, nil
}

// Reward queries info about a reward by using address and key
func (k Querier) Reward(
	c context.Context,
	req *types.QueryRewardRequest,
) (*types.QueryRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.LockerAddress)
	if err != nil {
		return nil, err
	}

	lock, err := k.GetLock(ctx, addr, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryRewardResponse{
		Reward: k.getReward(ctx, lock),
	}, nil
}

// Locks queries all locks with pagination.
func (k Querier) Locks(
	c context.Context,
	req *types.QueryLocksRequest,
) (*types.QueryLocksResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.LockerAddress)
	if err != nil {
		return nil, err
	}

	lockStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.LocksByAddressStoreKey(addr))

	filteredLocks, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		lockStore,
		req.Pagination,
		func(key []byte, s *types.Lock) (*types.LockResponse, error) {
			if !k.IsActiveKey(ctx, s.Key) || s.Amount.IsZero() {
				return nil, nil
			}

			return &types.LockResponse{
				Key:    s.Key,
				Amount: s.Amount,
			}, nil
		}, func() *types.Lock {
			return &types.Lock{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryLocksResponse{Locks: filteredLocks, Pagination: pageRes}, nil
}

// Lock queries info about a lock by using address and key
func (k Querier) Lock(
	c context.Context,
	req *types.QueryLockRequest,
) (*types.QueryLockResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.LockerAddress)
	if err != nil {
		return nil, err
	}

	isActive := k.IsActiveKey(ctx, req.Key)
	if !isActive {
		return nil, types.ErrKeyNotActive
	}

	lock, err := k.GetLock(ctx, addr, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryLockResponse{
		Lock: types.LockResponse{
			Key:    lock.Key,
			Amount: lock.Amount,
		},
	}, nil
}
