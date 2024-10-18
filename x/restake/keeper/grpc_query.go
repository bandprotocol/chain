package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper.
type Querier struct {
	Keeper
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

	vault, found := k.GetVault(ctx, req.Key)
	if !found {
		return nil, types.ErrVaultNotFound.Wrapf("key: %s", req.Key)
	}

	return &types.QueryVaultResponse{Vault: vault}, nil
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

	lock, found := k.GetLock(ctx, addr, req.Key)
	if !found {
		return nil, types.ErrLockNotFound.Wrapf("address: %s, key: %s", addr.String(), req.Key)
	}

	return &types.QueryLockResponse{
		Lock: types.LockResponse{
			Key:   lock.Key,
			Power: lock.Power,
		},
	}, nil
}

// Stake queries stake information of an address.
func (k Querier) Stake(
	c context.Context,
	req *types.QueryStakeRequest,
) (*types.QueryStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.StakerAddress)
	if err != nil {
		return nil, err
	}

	stake := k.GetStake(ctx, addr)
	return &types.QueryStakeResponse{Stake: stake}, nil
}

// Params queries all params of the module.
func (k Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: k.GetParams(ctx),
	}, nil
}
