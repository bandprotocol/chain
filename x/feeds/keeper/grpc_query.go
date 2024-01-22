package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	// convert filter symbols to map
	reqSymbols := make(map[string]bool, 0)
	for _, s := range req.Symbols {
		reqSymbols[s] = true
	}

	store := ctx.KVStore(q.keeper.storeKey)
	priceStore := prefix.NewStore(store, types.PriceStoreKeyPrefix)

	filteredPrices, pageRes, err := query.GenericFilteredPaginate(
		q.keeper.cdc,
		priceStore,
		req.Pagination,
		func(key []byte, p *types.Price) (*types.Price, error) {
			matchSymbol := true

			// match symbol
			if len(reqSymbols) != 0 {
				if _, ok := reqSymbols[p.Symbol]; !ok {
					matchSymbol = false
				}
			}

			if matchSymbol {
				return p, nil
			}

			return nil, nil
		}, func() *types.Price {
			return &types.Price{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPricesResponse{Prices: filteredPrices, Pagination: pageRes}, nil
}

func (q queryServer) Price(
	goCtx context.Context, req *types.QueryPriceRequest,
) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := q.keeper.GetSymbol(ctx, req.Symbol); err != nil {
		return nil, err
	}

	price, _ := q.keeper.GetPrice(ctx, req.Symbol)
	priceVals := q.keeper.GetPriceValidators(ctx, req.Symbol)

	return &types.QueryPriceResponse{
		Price:           price,
		PriceValidators: priceVals,
	}, nil
}

func (q queryServer) ValidatorPrices(
	goCtx context.Context, req *types.QueryValidatorPricesRequest,
) (*types.QueryValidatorPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	var priceVals []types.PriceValidator

	symbols := q.keeper.GetSymbols(ctx)
	for _, symbol := range symbols {
		priceVal, err := q.keeper.GetPriceValidator(ctx, symbol.Symbol, val)
		if err != nil {
			return nil, err
		}

		priceVals = append(priceVals, priceVal)
	}

	return &types.QueryValidatorPricesResponse{
		ValidatorPrices: priceVals,
	}, nil
}

func (q queryServer) PriceValidator(
	goCtx context.Context, req *types.QueryPriceValidatorRequest,
) (*types.QueryPriceValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	priceVal, err := q.keeper.GetPriceValidator(ctx, req.Symbol, val)
	if err != nil {
		return nil, err
	}

	return &types.QueryPriceValidatorResponse{
		PriceValidator: priceVal,
	}, nil
}

func (q queryServer) Symbols(
	goCtx context.Context, req *types.QuerySymbolsRequest,
) (*types.QuerySymbolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// convert filter symbols to map
	reqSymbols := make(map[string]bool, 0)
	for _, s := range req.Symbols {
		reqSymbols[s] = true
	}

	store := ctx.KVStore(q.keeper.storeKey)
	symbolStore := prefix.NewStore(store, types.SymbolStoreKeyPrefix)

	filteredSymbols, pageRes, err := query.GenericFilteredPaginate(
		q.keeper.cdc,
		symbolStore,
		req.Pagination,
		func(key []byte, s *types.Symbol) (*types.Symbol, error) {
			matchSymbol := true

			// match symbol
			if len(reqSymbols) != 0 {
				if _, ok := reqSymbols[s.Symbol]; !ok {
					matchSymbol = false
				}
			}

			if matchSymbol {
				return s, nil
			}

			return nil, nil
		}, func() *types.Symbol {
			return &types.Symbol{}
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySymbolsResponse{Symbols: filteredSymbols, Pagination: pageRes}, nil
}

func (q queryServer) OffChain(
	goCtx context.Context, req *types.QueryOffChainRequest,
) (*types.QueryOffChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryOffChainResponse{
		OffChain: q.keeper.GetOffChain(ctx),
	}, nil
}

func (q queryServer) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.keeper.GetParams(ctx),
	}, nil
}
