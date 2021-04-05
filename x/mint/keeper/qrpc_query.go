package keeper

import (
	"context"
	"github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (k Keeper) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (k Keeper) AnnualProvisions(c context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}

// EthIntegrationAddress returns ethereum integration address
func (k Keeper) EthIntegrationAddress(c context.Context, _ *types.QueryEthIntegrationAddressRequest) (*types.QueryEthIntegrationAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryEthIntegrationAddressResponse{EthIntegrationAddress: params.EthIntegrationAddress}, nil
}

// TreasuryPool returns current treasury pool
func (k Keeper) TreasuryPool(c context.Context, _ *types.QueryTreasuryPoolRequest) (*types.QueryTreasuryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	mintPool := k.GetMintPool(ctx)

	return &types.QueryTreasuryPoolResponse{TreasuryPool: mintPool.TreasuryPool}, nil
}
