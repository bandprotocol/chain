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

func (q queryServer) DelegatorSignals(
	goCtx context.Context, req *types.QueryDelegatorSignalsRequest,
) (*types.QueryDelegatorSignalsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	delegator, err := sdk.AccAddressFromBech32(req.Delegator)
	if err != nil {
		return nil, err
	}

	signals := q.keeper.GetDelegatorSignals(ctx, delegator)
	if signals == nil {
		return nil, status.Error(codes.Internal, "no signal")
	}
	return &types.QueryDelegatorSignalsResponse{Signals: signals}, nil
}

func (q queryServer) Prices(
	goCtx context.Context, req *types.QueryPricesRequest,
) (*types.QueryPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// convert filter signal ids to map
	reqSignalIDs := make(map[string]bool)
	for _, s := range req.SignalIds {
		reqSignalIDs[s] = true
	}

	store := ctx.KVStore(q.keeper.storeKey)
	priceStore := prefix.NewStore(store, types.PriceStoreKeyPrefix)

	filteredPrices, pageRes, err := query.GenericFilteredPaginate(
		q.keeper.cdc,
		priceStore,
		req.Pagination,
		func(key []byte, p *types.Price) (*types.Price, error) {
			matchSignalID := true

			// match signal id
			if len(reqSignalIDs) != 0 {
				if _, ok := reqSignalIDs[p.SignalID]; !ok {
					matchSignalID = false
				}
			}

			if matchSignalID {
				return p, nil
			}

			return nil, nil
		}, func() *types.Price {
			return &types.Price{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPricesResponse{Prices: filteredPrices, Pagination: pageRes}, nil
}

func (q queryServer) Price(
	goCtx context.Context, req *types.QueryPriceRequest,
) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	f, err := q.keeper.GetFeed(ctx, req.SignalId)
	if err != nil {
		return nil, err
	}

	price, _ := q.keeper.GetPrice(ctx, req.SignalId)
	priceVals := q.keeper.GetPriceValidators(ctx, req.SignalId)

	var filteredPriceVals []types.PriceValidator
	blockTime := ctx.BlockTime().Unix()
	for _, priceVal := range priceVals {
		if priceVal.Timestamp > blockTime-f.Interval {
			filteredPriceVals = append(filteredPriceVals, priceVal)
		}
	}

	return &types.QueryPriceResponse{
		Price:           price,
		PriceValidators: filteredPriceVals,
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

	feeds := q.keeper.GetFeeds(ctx)
	for _, feed := range feeds {
		priceVal, err := q.keeper.GetPriceValidator(ctx, feed.SignalID, val)
		if err == nil {
			priceVals = append(priceVals, priceVal)
		}
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

	priceVal, err := q.keeper.GetPriceValidator(ctx, req.SignalId, val)
	if err != nil {
		return nil, err
	}

	return &types.QueryPriceValidatorResponse{
		PriceValidator: priceVal,
	}, nil
}

func (q queryServer) ValidValidator(
	goCtx context.Context, req *types.QueryValidValidatorRequest,
) (*types.QueryValidValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	flag := true

	// check if it's in top bonded validators.
	isTop := q.keeper.IsTopValidator(ctx, req.Validator)
	if !isTop {
		flag = false
	}

	validatorStatus := q.keeper.oracleKeeper.GetValidatorStatus(ctx, val)
	if !validatorStatus.IsActive {
		flag = false
	}

	return &types.QueryValidValidatorResponse{Valid: flag}, nil
}

func (q queryServer) Feeds(
	goCtx context.Context, req *types.QueryFeedsRequest,
) (*types.QueryFeedsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// convert filter signal ids to map
	reqSignalIDs := make(map[string]bool)
	for _, s := range req.SignalIds {
		reqSignalIDs[s] = true
	}

	store := ctx.KVStore(q.keeper.storeKey)
	feedStore := prefix.NewStore(store, types.FeedStoreKeyPrefix)

	filteredFeeds, pageRes, err := query.GenericFilteredPaginate(
		q.keeper.cdc,
		feedStore,
		req.Pagination,
		func(key []byte, f *types.Feed) (*types.Feed, error) {
			matchSignalID := true

			// match signal id
			if len(reqSignalIDs) != 0 {
				if _, ok := reqSignalIDs[f.SignalID]; !ok {
					matchSignalID = false
				}
			}

			if matchSignalID {
				return f, nil
			}

			return nil, nil
		}, func() *types.Feed {
			return &types.Feed{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFeedsResponse{Feeds: filteredFeeds, Pagination: pageRes}, nil
}

func (q queryServer) SupportedFeeds(
	goCtx context.Context, _ *types.QuerySupportedFeedsRequest,
) (*types.QuerySupportedFeedsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QuerySupportedFeedsResponse{
		Feeds: q.keeper.GetSupportedFeedsByPower(ctx),
	}, nil
}

func (q queryServer) PriceService(
	goCtx context.Context, _ *types.QueryPriceServiceRequest,
) (*types.QueryPriceServiceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryPriceServiceResponse{
		PriceService: q.keeper.GetPriceService(ctx),
	}, nil
}

func (q queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.keeper.GetParams(ctx),
	}, nil
}
