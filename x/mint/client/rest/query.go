package rest

import (
	"fmt"
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		fmt.Sprintf("%s/%s", minttypes.ModuleName, minttypes.QueryParams),
		queryParamsHandlerFn(clientCtx),
	).Methods("GET")

	r.HandleFunc(
		fmt.Sprintf("%s/%s", minttypes.ModuleName, minttypes.QueryInflation),
		queryInflationHandlerFn(clientCtx),
	).Methods("GET")

	r.HandleFunc(
		fmt.Sprintf("%s/%s", minttypes.ModuleName, minttypes.QueryAnnualProvisions),
		queryAnnualProvisionsHandlerFn(clientCtx),
	).Methods("GET")

	r.HandleFunc(
		fmt.Sprintf("%s/%s", minttypes.ModuleName, minttypes.QueryEthIntegrationAddress),
		queryEthIntegrationAddressHandlerFn(clientCtx),
	).Methods("GET")

	r.HandleFunc(
		fmt.Sprintf("%s/%s", minttypes.ModuleName, minttypes.QueryTreasuryPool),
		queryTreasuryPoolHandlerFn(clientCtx),
	).Methods("GET")
}

func queryParamsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", minttypes.QuerierRoute, minttypes.QueryParams)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func queryInflationHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", minttypes.QuerierRoute, minttypes.QueryInflation)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func queryAnnualProvisionsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", minttypes.QuerierRoute, minttypes.QueryAnnualProvisions)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func queryEthIntegrationAddressHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", minttypes.QuerierRoute, minttypes.QueryEthIntegrationAddress)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryTreasuryPoolHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", minttypes.QuerierRoute, minttypes.QueryTreasuryPool)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(route)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
