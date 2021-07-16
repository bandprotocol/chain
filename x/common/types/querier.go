package types

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	"net/http"
)

// QueryResult wraps querier result with HTTP status to return to application.
type QueryResult struct {
	Status int             `json:"status"`
	Result json.RawMessage `json:"result"`
}

// QueryOK creates and marshals a QueryResult instance with HTTP status OK.
func QueryOK(legacyQuerierCdc *codec.LegacyAmino, result interface{}) ([]byte, error) {
	return codec.MarshalJSONIndent(legacyQuerierCdc, QueryResult{
		Status: http.StatusOK,
		Result: codec.MustMarshalJSONIndent(legacyQuerierCdc, result),
	})
}

// QueryBadRequest creates and marshals a QueryResult instance with HTTP status BadRequest.
func QueryBadRequest(legacyQuerierCdc *codec.LegacyAmino, result interface{}) ([]byte, error) {
	return codec.MarshalJSONIndent(legacyQuerierCdc, QueryResult{
		Status: http.StatusBadRequest,
		Result: codec.MustMarshalJSONIndent(legacyQuerierCdc, result),
	})
}

// QueryNotFound creates and marshals a QueryResult instance with HTTP status NotFound.
func QueryNotFound(legacyQuerierCdc *codec.LegacyAmino, result interface{}) ([]byte, error) {
	return codec.MarshalJSONIndent(legacyQuerierCdc, QueryResult{
		Status: http.StatusBadRequest,
		Result: codec.MustMarshalJSONIndent(legacyQuerierCdc, result),
	})
}

type QueryPaginationParams struct {
	Offset uint64 `json:"offset" yaml:"offset"`
	Limit  uint64 `json:"limit" yaml:"limit"`
	Desc   bool   `json:"desc" yaml:"desc"`
}
