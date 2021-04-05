package rest

import (
	"fmt"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"net/http"
)

type RequestPrices struct {
	Symbols  []string `json:"symbols"`
	MinCount uint64   `json:"min_count"`
	AskCount uint64   `json:"ask_count"`
}

func getParamsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryParams))
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getCountsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryCounts))
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// todo add data
//func getDataByHashHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
//		if !ok {
//			return
//		}
//		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryData), nil)
//		if rest.CheckInternalServerError(w, err) {
//			return
//		}
//
//		clientCtx = clientCtx.WithHeight(height)
//		rest.PostProcessResponse(w, clientCtx, res)
//	}
//}
//
//func getDataSourceByIDHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		vars := mux.Vars(r)
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, oracletypes.QueryDataSources, vars[idTag]))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getOracleScriptByIDHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		vars := mux.Vars(r)
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, oracletypes.QueryOracleScripts, vars[idTag]))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getRequestByIDHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		vars := mux.Vars(r)
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, oracletypes.QueryRequests, vars[idTag]))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		bz, height, err := clientcmn.QuerySearchLatestRequest(
//			route, cliCtx,
//			r.FormValue("oid"), r.FormValue("calldata"), r.FormValue("ask_count"), r.FormValue("min_count"),
//		)
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getRequestsPricesHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		decoder := json.NewDecoder(r.Body)
//		var requestPrices RequestPrices
//		err := decoder.Decode(&requestPrices)
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
//			return
//		}
//		prices := make([]price.Price, len(requestPrices.Symbols))
//		height := int64(0)
//		for idx, symbol := range requestPrices.Symbols {
//			bz, h, err := cliCtx.Query(fmt.Sprintf("band/prices/%s/%d/%d", symbol, requestPrices.AskCount, requestPrices.MinCount))
//			if h > height {
//				height = h
//			}
//			if err != nil {
//				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
//				return
//			}
//			var price price.Price
//			err = cliCtx.Codec.UnmarshalBinaryBare(bz, &price)
//			if err != nil {
//				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//				return
//			}
//			prices[idx] = price
//		}
//		bz, err := commontypes.QueryOK(oracletypes.ModuleCdc, prices)
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getRequestsPriceSymbolsHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//
//		bz, height, err := cliCtx.Query(fmt.Sprintf("band/price_symbols/%s/%s", r.FormValue("ask_count"), r.FormValue("min_count")))
//
//		var symbols []string
//		if err := cliCtx.Codec.UnmarshalBinaryBare(bz, &symbols); err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		bz, err = commontypes.QueryOK(oracletypes.ModuleCdc, symbols)
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getMultiRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		limit := 1
//		if rawLimit := r.FormValue("limit"); rawLimit != "" {
//			var err error
//			limit, err = strconv.Atoi(rawLimit)
//			if err != nil {
//				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
//			}
//		}
//		bz, height, err := clientcmn.QueryMultiSearchLatestRequest(
//			route, cliCtx,
//			r.FormValue("oid"), r.FormValue("calldata"), r.FormValue("ask_count"), r.FormValue("min_count"), limit,
//		)
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getValidatorStatusHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		vars := mux.Vars(r)
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, oracletypes.QueryValidatorStatus, vars[validatorAddressTag]))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getReportersHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		vars := mux.Vars(r)
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", route, oracletypes.QueryReporters, vars[validatorAddressTag]))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func getActiveValidatorsHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", route, oracletypes.QueryActiveValidators))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
//
//func dataProvidersPoolHandler(clientCtx client.Context) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
//		if !ok {
//			return
//		}
//		bz, height, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, oracletypes.QueryDataProvidersPool))
//		if err != nil {
//			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
//			return
//		}
//		commonrest.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
//	}
//}
