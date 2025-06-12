package client

import "github.com/bandprotocol/chain/v3/x/tss/types"

// DEResponse wraps the types.QueryDEResponse to provide additional helper methods.
type DEResponse struct {
	types.QueryDEResponse
	blockHeight int64
}

// NewDEResponse creates a new instance of DEResponse.
func NewDEResponse(der *types.QueryDEResponse, blockHeight int64) *DEResponse {
	return &DEResponse{*der, blockHeight}
}

// GetRemaining retrieves the remaining DE in the blockchain for the address.
func (der DEResponse) GetRemaining() uint64 {
	return der.Pagination.Total
}

func (der DEResponse) GetBlockHeight() int64 {
	return der.blockHeight
}
