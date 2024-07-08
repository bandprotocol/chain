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

// DelegatorSignals queries all signals submitted by a delegator.
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
		return nil, status.Error(codes.NotFound, "no signal")
	}
	return &types.QueryDelegatorSignalsResponse{Signals: signals}, nil
}

// Prices queries all current prices.
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

// Price queries price of a signal id.
func (q queryServer) Price(
	goCtx context.Context, req *types.QueryPriceRequest,
) (*types.QueryPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	price, err := q.keeper.GetPrice(ctx, req.SignalId)
	if err != nil {
		return &types.QueryPriceResponse{}, err
	}

	return &types.QueryPriceResponse{
		Price: price,
	}, nil
}

// ValidatorPrices queries all price-validator submitted by a validator.
func (q queryServer) ValidatorPrices(
	goCtx context.Context, req *types.QueryValidatorPricesRequest,
) (*types.QueryValidatorPricesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	valPricesList, err := q.keeper.GetValidatorPriceList(ctx, val)
	if err != nil {
		return &types.QueryValidatorPricesResponse{
			ValidatorPrices: []types.ValidatorPrice{},
		}, nil
	}

	if len(req.SignalIds) == 0 {
		// Return all validator prices if SignalIds is empty
		return &types.QueryValidatorPricesResponse{
			ValidatorPrices: valPricesList.ValidatorPrices,
		}, nil
	}

	// Filter validator prices based on requested SignalIds
	filteredPrices := make([]types.ValidatorPrice, 0, len(req.SignalIds))
	signalIDSet := make(map[string]struct{})
	for _, id := range req.SignalIds {
		signalIDSet[id] = struct{}{}
	}

	for _, valPrice := range valPricesList.ValidatorPrices {
		if _, exists := signalIDSet[valPrice.SignalID]; exists {
			filteredPrices = append(filteredPrices, valPrice)
		}
	}

	return &types.QueryValidatorPricesResponse{
		ValidatorPrices: filteredPrices,
	}, nil
}

// ValidValidator queries whether a validator is required to send price.
func (q queryServer) ValidValidator(
	goCtx context.Context, req *types.QueryValidValidatorRequest,
) (*types.QueryValidValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}

	isValid := true

	// check if it's bonded validators.
	isBonded := q.keeper.IsBondedValidator(ctx, val)
	if !isBonded {
		isValid = false
	}

	validatorStatus := q.keeper.oracleKeeper.GetValidatorStatus(ctx, val)
	if !validatorStatus.IsActive {
		isValid = false
	}

	return &types.QueryValidValidatorResponse{Valid: isValid}, nil
}

// SignalTotalPowers queries all current signal-total-powers.
func (q queryServer) SignalTotalPowers(
	goCtx context.Context, req *types.QuerySignalTotalPowersRequest,
) (*types.QuerySignalTotalPowersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// convert filter signal ids to map
	reqSignalIDs := make(map[string]bool)
	for _, s := range req.SignalIds {
		reqSignalIDs[s] = true
	}

	store := ctx.KVStore(q.keeper.storeKey)
	signalTotalPowerStore := prefix.NewStore(store, types.SignalTotalPowerStoreKeyPrefix)

	filteredSignalTotalPowers, pageRes, err := query.GenericFilteredPaginate(
		q.keeper.cdc,
		signalTotalPowerStore,
		req.Pagination,
		func(key []byte, s *types.Signal) (*types.Signal, error) {
			matchSignalID := true

			// match signal id
			if len(reqSignalIDs) != 0 {
				if _, ok := reqSignalIDs[s.ID]; !ok {
					matchSignalID = false
				}
			}

			if matchSignalID {
				return s, nil
			}

			return nil, nil
		}, func() *types.Signal {
			return &types.Signal{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySignalTotalPowersResponse{
		SignalTotalPowers: filteredSignalTotalPowers,
		Pagination:        pageRes,
	}, nil
}

// SupportedFeeds queries all current supported feed-with-deviations.
func (q queryServer) SupportedFeeds(
	goCtx context.Context, _ *types.QuerySupportedFeedsRequest,
) (*types.QuerySupportedFeedsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supportedFeeds := q.keeper.GetSupportedFeeds(ctx)
	feedWithDeviations := make([]types.FeedWithDeviation, 0, len(supportedFeeds.Feeds))
	for _, feed := range supportedFeeds.Feeds {
		deviation := CalculateDeviation(feed.Power, q.keeper.GetParams(ctx))
		feedWithDeviations = append(
			feedWithDeviations,
			types.FeedWithDeviation{
				SignalID:            feed.SignalID,
				Power:               feed.Power,
				Interval:            feed.Interval,
				DeviationBasisPoint: deviation,
			})
	}

	return &types.QuerySupportedFeedsResponse{
		SupportedFeeds: types.SupportedFeedWithDeviations{
			Feeds:               feedWithDeviations,
			LastUpdateTimestamp: supportedFeeds.LastUpdateTimestamp,
			LastUpdateBlock:     supportedFeeds.LastUpdateBlock,
		},
	}, nil
}

// ReferenceSourceConfig queries current reference source config.
func (q queryServer) ReferenceSourceConfig(
	goCtx context.Context, _ *types.QueryReferenceSourceConfigRequest,
) (*types.QueryReferenceSourceConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryReferenceSourceConfigResponse{
		ReferenceSourceConfig: q.keeper.GetReferenceSourceConfig(ctx),
	}, nil
}

// Params queries all params of feeds module.
func (q queryServer) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{
		Params: q.keeper.GetParams(ctx),
	}, nil
}

// IsFeeder queries if the given address is a feeder grantee of the validator
func (q queryServer) IsFeeder(
	c context.Context,
	req *types.QueryIsFeederRequest,
) (*types.QueryIsFeederResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	val, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	feeder, err := sdk.AccAddressFromBech32(req.FeederAddress)
	if err != nil {
		return nil, err
	}
	return &types.QueryIsFeederResponse{IsFeeder: q.keeper.IsFeeder(ctx, val, feeder)}, nil
}