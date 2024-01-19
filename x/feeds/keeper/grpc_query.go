package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	keeper Keeper
}

func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{
		keeper: k,
	}
}

func (q queryServer) Prices(
	goCtx context.Context, req *types.QueryPricesRequest,
) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO:
	// - add filter
	// - add pagination

	return &types.QueryPricesResponse{
		Prices: q.keeper.GetPrices(ctx),
	}, nil
}

func (q queryServer) Price(
	goCtx context.Context, req *types.QueryPriceRequest,
) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: handle error
	price, _ := q.keeper.GetPrice(ctx, req.Symbol)
	priceVals := q.keeper.GetPriceValidators(ctx, req.Symbol)

	return &types.QueryPriceResponse{
		Price:           price,
		PriceValidators: priceVals,
	}, nil
}

func (q queryServer) Symbols(
	goCtx context.Context, req *types.QuerySymbolsRequest,
) (*types.QuerySymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO
	// - add pagination
	// - add filter

	return &types.QuerySymbolsResponse{
		Symbols: q.keeper.GetSymbols(ctx),
	}, nil
}

func (q queryServer) OffChain(
	goCtx context.Context, req *types.QueryOffChainRequest,
) (*types.QueryOffChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	offChain, err := q.keeper.GetOffChain(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryOffChainResponse{
		OffChain: offChain,
	}, nil
}

func (q queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.keeper.GetParams(ctx),
	}, nil
}
