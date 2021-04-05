package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the oracle Querier.
const (
	QueryParams             = "params"
	QueryCounts             = "counts"
	QueryData               = "data"
	QueryDataSources        = "data_sources"
	QueryOracleScripts      = "oracle_scripts"
	QueryRequests           = "requests"
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

// QueryCountsResult is the struct for the result of query counts.
type QueryCountsResult struct {
	DataSourceCount   int64 `json:"data_source_count"`
	OracleScriptCount int64 `json:"oracle_script_count"`
	RequestCount      int64 `json:"request_count"`
}

// QueryRequestResult is the struct for the result of request query.
type QueryRequestResult struct {
	Request Request  `json:"request"`
	Reports []Report `json:"reports"`
	Result  *Result  `json:"result"`
}

// QueryActiveValidatorResult is the struct for the result of request active validators.
type QueryActiveValidatorResult struct {
	Address sdk.ValAddress `json:"address"`
	Power   uint64         `json:"power"`
}

type QueryRequestSearchParams struct {
	OracleScriptID OracleScriptID `json:"oracle_script_id" yaml:"oracle_script_id"`
	CallData       []byte         `json:"call_data" yaml:"call_data"`
	AskCount       int64          `json:"ask_count" yaml:"ask_count"`
	MinCount       int64          `json:"min_count" yaml:"min_count"`
}

type RequestPrices struct {
	Symbols  []string `json:"symbols"`
	MinCount uint64   `json:"min_count"`
	AskCount uint64   `json:"ask_count"`
}

type QueryRequestPricesParams struct {
	Symbol   string `json:"symbol" yaml:"symbol"`
	MinCount uint64 `json:"min_count" yaml:"min_count"`
	AskCount uint64 `json:"ask_count" yaml:"ask_count"`
}

func NewQueryRequestSearchParams(oid OracleScriptID, callData []byte, askCount, minCount int64) QueryRequestSearchParams {
	return QueryRequestSearchParams{
		OracleScriptID: oid,
		CallData:       callData,
		AskCount:       askCount,
		MinCount:       minCount,
	}
}

func NewQueryRequestPricesParams(symbol string, minCount, askCount uint64) QueryRequestPricesParams {
	return QueryRequestPricesParams{
		Symbol:   symbol,
		MinCount: minCount,
		AskCount: askCount,
	}
}
