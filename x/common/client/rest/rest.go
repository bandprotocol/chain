package rest

import (
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func СheckPaginationParams(w http.ResponseWriter, r *http.Request) (commontypes.QueryPaginationParams, bool) {
	vars := mux.Vars(r)
	limit, err := strconv.ParseUint(vars[LimitTag], 10, 64)
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}
	offset, err := strconv.ParseUint(vars[OffsetTag], 10, 64)
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}
	desc, err := strconv.ParseBool(vars[DescTag])
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}

	return commontypes.QueryPaginationParams{Offset: offset, Limit: limit, Desc: desc}, true
}
