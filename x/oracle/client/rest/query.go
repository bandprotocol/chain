package rest

import (
	"fmt"
	hookscommon "github.com/GeoDB-Limited/odin-core/hooks/common"
	hookprice "github.com/GeoDB-Limited/odin-core/hooks/price"
	commontypes "github.com/GeoDB-Limited/odin-core/x/common/types"
	oracleclientcommon "github.com/GeoDB-Limited/odin-core/x/oracle/client/common"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

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

func getDataByHashHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryData, vars[dataHashTag]))
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getDataSourceByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryDataSources, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getOracleScriptByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryOracleScripts, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getRequestByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		res, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryRequests, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		oid, err := strconv.ParseInt(r.FormValue("oid"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		callData := []byte(r.FormValue("calldata"))

		askCount, err := strconv.ParseInt(r.FormValue("ask_count"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		minCount, err := strconv.ParseInt(r.FormValue("min_count"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO add search endpoint to querier
		bin := clientCtx.LegacyAmino.MustMarshalJSON(oracletypes.NewQueryRequestSearchParams(oracletypes.OracleScriptID(oid), callData, askCount, minCount))
		res, height, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryRequests), bin)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

// TODO fix later
func getRequestsPricesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPrices oracletypes.RequestPrices

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &requestPrices) {
			return
		}

		prices := make([]hookprice.Price, len(requestPrices.Symbols))
		height := int64(0)
		for idx, symbol := range requestPrices.Symbols {

			bin := clientCtx.LegacyAmino.MustMarshalJSON(oracletypes.NewQueryRequestPricesParams(symbol, requestPrices.MinCount, requestPrices.AskCount))
			res, h, err := clientCtx.QueryWithData(fmt.Sprintf("%s/%s", hookscommon.AppHook, oracletypes.QueryRequestPrices), bin)
			if h > height {
				height = h
			}
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			var price hookprice.Price
			err = clientCtx.LegacyAmino.UnmarshalBinaryBare(res, &price)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			prices[idx] = price
		}

		bz, err := commontypes.QueryOK(clientCtx.LegacyAmino, prices)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

// TODO: fix later
func getRequestsPriceSymbolsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		bz, height, err := clientCtx.Query(fmt.Sprintf("%s/%s", hookscommon.AppHook, oracletypes.QueryPriceSymbols))

		var symbols []string
		if err := clientCtx.LegacyAmino.UnmarshalBinaryBare(bz, &symbols); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		bz, err = commontypes.QueryOK(clientCtx.LegacyAmino, symbols)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getMultiRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		// TODO: maybe use rest.ParseHTTPArgsWithLimit
		limit := 1
		if rawLimit := r.FormValue("limit"); rawLimit != "" {
			var err error
			limit, err = strconv.Atoi(rawLimit)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			}
		}

		oid, err := strconv.ParseInt(r.FormValue("oid"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		callData := []byte(r.FormValue("calldata"))

		askCount, err := strconv.ParseInt(r.FormValue("ask_count"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		minCount, err := strconv.ParseInt(r.FormValue("min_count"), 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bz, height, err := oracleclientcommon.QueryMultiSearchLatestRequest(
			oracletypes.QuerierRoute, clientCtx,
			oracletypes.NewQueryRequestSearchParams(oracletypes.OracleScriptID(oid), callData, askCount, minCount), limit,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getValidatorStatusHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryValidatorStatus, vars[validatorAddressTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getReportersHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryReporters, vars[validatorAddressTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getActiveValidatorsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryActiveValidators))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func dataProvidersPoolHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryDataProvidersPool))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}
