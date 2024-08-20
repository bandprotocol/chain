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

// Vaults queries all vaults with pagination.
func (k Querier) Vaults(
	c context.Context,
	req *types.QueryVaultsRequest,
) (*types.QueryVaultsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	vaultStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.VaultStoreKeyPrefix)

	filteredVaults, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		vaultStore,
		req.Pagination,
		func(key []byte, v *types.Vault) (*types.Vault, error) {
			return v, nil
		}, func() *types.Vault {
			return &types.Vault{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVaultsResponse{Vaults: filteredVaults, Pagination: pageRes}, nil
}

// Vault queries info about a vault.
func (k Querier) Vault(
	c context.Context,
	req *types.QueryVaultRequest,
) (*types.QueryVaultResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	vault, err := k.GetVault(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryVaultResponse{Vault: vault}, nil
}

// Rewards queries all rewards with pagination.
func (k Querier) Rewards(
	c context.Context,
	req *types.QueryRewardsRequest,
) (*types.QueryRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.StakerAddress)
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

	addr, err := sdk.AccAddressFromBech32(req.StakerAddress)
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

	addr, err := sdk.AccAddressFromBech32(req.StakerAddress)
	if err != nil {
		return nil, err
	}

	lockStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.LocksByAddressStoreKey(addr))

	filteredLocks, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		lockStore,
		req.Pagination,
		func(key []byte, s *types.Lock) (*types.LockResponse, error) {
			if !k.IsActiveVault(ctx, s.Key) || s.Power.IsZero() {
				return nil, nil
			}

			return &types.LockResponse{
				Key:   s.Key,
				Power: s.Power,
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

	addr, err := sdk.AccAddressFromBech32(req.StakerAddress)
	if err != nil {
		return nil, err
	}

	isActive := k.IsActiveVault(ctx, req.Key)
	if !isActive {
		return nil, types.ErrVaultNotActive
	}

	lock, err := k.GetLock(ctx, addr, req.Key)
	if err != nil {
		return nil, err
	}

	return &types.QueryLockResponse{
		Lock: types.LockResponse{
			Key:   lock.Key,
			Power: lock.Power,
		},
	}, nil
}
