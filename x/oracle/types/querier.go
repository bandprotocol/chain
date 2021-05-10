package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Query endpoints supported by the oracle Querier.
const (
	QueryParams             = "params"
	QueryCounts             = "counts"
	QueryData               = "data"
	QueryDataSources        = "data_sources"
	QueryOracleScripts      = "oracle_scripts"
	QueryRequests           = "requests"
	QueryRequestReports     = "request_reports"
	QueryRequestPrices      = "request_prices"
	QueryPriceSymbols       = "price_symbols"
	QueryLatestRequest      = "latest_request"
	QueryMultiRequestSearch = "multi_request_search"
	QueryRequestSearch      = "request_search"
	QueryValidatorStatus    = "validator_status"
	QueryReporters          = "reporters"
	QueryActiveValidators   = "active_validators"
	QueryPendingRequests    = "pending_requests"
	QueryDataProvidersPool  = "data_providers_pool"
)

// deprecated: for legacy REST use only
// QueryActiveValidatorResult is the struct for the result of request active validators.
type QueryActiveValidatorResult struct {
	Address sdk.ValAddress `json:"address"`
	Power   uint64         `json:"power"`
}

// deprecated: for legacy REST use only
// QueryRequestResult is the struct for the result of request query.
type QueryRequestResult struct {
	Request Request  `json:"request"`
	Reports []Report `json:"reports"`
	Result  *Result  `json:"result"`
}

func NewQueryRequestSearchRequest(oid int64, callData []byte, askCount, minCount int64) *QueryRequestSearchRequest {
	return &QueryRequestSearchRequest{
		OracleScriptId: oid,
		Calldata:       callData,
		AskCount:       askCount,
		MinCount:       minCount,
	}
}

func NewQueryRequestSearchResponse(req QueryRequestResponse) *QueryRequestSearchResponse {
	return &QueryRequestSearchResponse{
		RequestPacketData:  req.Request.RequestPacketData,
		ResponsePacketData: req.Request.ResponsePacketData,
	}
}

func NewQueryRequestPricesRequest(symbol string, minCount, askCount int64) QueryRequestPriceRequest {
	return QueryRequestPriceRequest{
		Symbol:   symbol,
		MinCount: minCount,
		AskCount: askCount,
	}
}

type QueryPaginationParams struct {
	Offset uint64 `json:"offset" yaml:"offset"`
	Limit  uint64 `json:"limit" yaml:"limit"`
}
