package keeper

import (
	"context"

	"github.com/bandprotocol/chain/x/oracle/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

// Counts queries the number of data sources, oracle scripts, and requests.
func (k Querier) Counts(c context.Context, req *types.QueryCountsRequest) (*types.QueryCountsResponse, error) {
	return &types.QueryCountsResponse{}, nil
}

// Data queries the data source or oracle script script for given file hash.
func (k Querier) Data(c context.Context, req *types.QueryDataRequest) (*types.QueryDataResponse, error) {
	return &types.QueryDataResponse{}, nil
}

// DataSource queries data source info for given data source id.
func (k Querier) DataSource(c context.Context, req *types.QueryDataSourceRequest) (*types.QueryDataSourceResponse, error) {
	return &types.QueryDataSourceResponse{}, nil
}

// OracleScript queries oracle script info for given oracle script id.
func (k Querier) OracleScript(c context.Context, req *types.QueryOracleScriptRequest) (*types.QueryOracleScriptResponse, error) {
	return &types.QueryOracleScriptResponse{}, nil
}

// Request queries request info for given request id.
func (k Querier) Request(c context.Context, req *types.QueryRequestRequest) (*types.QueryRequestResponse, error) {
	return &types.QueryRequestResponse{}, nil
}

// Validator queries oracle info of validator for given validator
// address.
func (k Querier) Validator(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	return &types.QueryValidatorResponse{}, nil
}

// Reporters queries all reporters of a given validator address.
func (k Querier) Reporters(c context.Context, req *types.QueryReportersRequest) (*types.QueryReportersResponse, error) {
	return &types.QueryReportersResponse{}, nil
}

// ActiveValidators queries all active oracle validators.
func (k Querier) ActiveValidators(c context.Context, req *types.QueryActiveValidatorsRequest) (*types.QueryActiveValidatorsResponse, error) {
	return &types.QueryActiveValidatorsResponse{}, nil
}

// Params queries the oracle parameters.
func (k Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{}, nil
}

// RequestSearch queries the latest request that match the given input.
func (k Querier) RequestSearch(c context.Context, req *types.QueryRequestSearchRequest) (*types.QueryRequestSearchResponse, error) {
	return &types.QueryRequestSearchResponse{}, nil
}

// RequestPrice queries the latest price on standard price reference oracle
// script.
func (k Querier) RequestPrice(c context.Context, req *types.QueryRequestPriceRequest) (*types.QueryRequestPriceResponse, error) {
	return &types.QueryRequestPriceResponse{}, nil
}
