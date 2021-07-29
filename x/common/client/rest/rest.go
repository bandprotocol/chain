package rest

import (
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"net/http"
	"strconv"
)

func EmptyOrDefault(val, defVal string) string {
	if len(val) == 0 {
		return defVal
	}
	return val
}

func СheckPaginationParams(w http.ResponseWriter, r *http.Request) (commontypes.QueryPaginationParams, bool) {
	urlQuery := r.URL.Query()
	limit, err := strconv.ParseUint(EmptyOrDefault(urlQuery.Get(LimitTag), strconv.Itoa(rest.DefaultLimit)), 10, 64)
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}
	offset, err := strconv.ParseUint(EmptyOrDefault(urlQuery.Get(OffsetTag), "0"), 10, 64)
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}
	countTotal, err := strconv.ParseBool(EmptyOrDefault(urlQuery.Get(CountTotalTag), "false"))
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}
	desc, err := strconv.ParseBool(EmptyOrDefault(urlQuery.Get(DescTag), "false"))
	if rest.CheckBadRequestError(w, err) {
		return commontypes.QueryPaginationParams{}, false
	}

	return commontypes.QueryPaginationParams{PageRequest: query.PageRequest{Offset: offset, Limit: limit, CountTotal: countTotal}, Desc: desc}, true
}
