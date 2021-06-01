package common

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

func PostProcessQueryResponse(w http.ResponseWriter, cliCtx client.Context, bz []byte) {
	var result types.QueryResult
	if err := json.Unmarshal(bz, &result); err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Status)
	rest.PostProcessResponse(w, cliCtx, result.Result)
}

func ValueOrDefault(val string, def interface{}) interface{} {
	if val == "" {
		return def
	}
	return val
}
