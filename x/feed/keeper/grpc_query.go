package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feed/types"
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
	// if err != nil {
	// 	return nil, err
	// }

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

	return &types.QuerySymbolsResponse{
		Symbols: q.keeper.GetSymbols(ctx),
	}, nil
}

func (q queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.keeper.GetParams(ctx),
	}, nil
}
