package client

import "github.com/bandprotocol/chain/v2/x/tss/types"

// DEResponse wraps the types.QueryDEResponse to provide additional helper methods.
type DEResponse struct {
	types.QueryDEResponse
}

// NewDEResponse creates a new instance of DEResponse.
func NewDEResponse(der *types.QueryDEResponse) *DEResponse {
	return &DEResponse{*der}
}

// GetRemaining retrieves the remaining DE in the blockchain for the address.
func (der DEResponse) GetRemaining() uint64 {
	return der.Pagination.Total
}
