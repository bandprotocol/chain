package rest

import (
	"fmt"
	commonrest "github.com/GeoDB-Limited/odin-core/x/common/client/rest"
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterRoutes TODO: refactor rest limit and offset params to get params
func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", telemetrytypes.ModuleName, telemetrytypes.QueryTopBalances), getTopBalancesHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", telemetrytypes.ModuleName, telemetrytypes.QueryExtendedValidators), getExtendedValidatorsHandler(clientCtx)).Methods("GET")
}

func getTopBalancesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		paginationParams, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bin := clientCtx.LegacyAmino.MustMarshalJSON(paginationParams)

		query := r.URL.Query()

		res, height, err := clientCtx.QueryWithData(fmt.Sprintf(
			"custom/%s/%s/%s",
			telemetrytypes.QuerierRoute,
			telemetrytypes.QueryTopBalances,
			query.Get(telemetrytypes.DenomTag),
		), bin)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getExtendedValidatorsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		paginationParams, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bin := clientCtx.LegacyAmino.MustMarshalJSON(paginationParams)

		vars := mux.Vars(r)

		res, height, err := clientCtx.QueryWithData(fmt.Sprintf(
			"custom/%s/%s/%s",
			telemetrytypes.QuerierRoute,
			telemetrytypes.QueryExtendedValidators,
			vars[telemetrytypes.StatusTag],
		), bin)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
