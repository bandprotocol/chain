package keeper

import (
	"context"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ telemetrytypes.QueryServer = Keeper{}

func (k Keeper) TopBalances(c context.Context, request *telemetrytypes.QueryTopBalancesRequest) (*telemetrytypes.QueryTopBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	balances, total := k.GetPaginatedBalances(ctx, request.GetDenom(), request.GetDesc(), request.Pagination)
	return &telemetrytypes.QueryTopBalancesResponse{
		Balances: balances,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}, nil
}
