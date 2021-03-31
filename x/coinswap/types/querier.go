package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the coinswap Querier.
const (
	QueryParams = "params"
	QueryRate   = "rate"
)

type QueryRateResult struct {
	Rate        sdk.Dec `json:"rate"`
	InitialRate sdk.Dec `json:"initial_rate"`
}
