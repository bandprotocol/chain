package rest

import (
	"fmt"
	hookscommon "github.com/GeoDB-Limited/odin-core/hooks/common"
	hookprice "github.com/GeoDB-Limited/odin-core/hooks/price"
	commonrest "github.com/GeoDB-Limited/odin-core/x/common/client/rest"
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
		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryParams,
		))
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
		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryCounts,
		))
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

		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryData,
			vars[dataHashTag],
		))
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

		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryDataSources,
			vars[idTag],
		))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getDataSourcesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		res, height, err := clientCtx.QueryWithData(
			fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryDataSources),
			bz,
		)
		if rest.CheckInternalServerError(w, err) {
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
		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryOracleScripts,
			vars[idTag],
		))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getOracleScriptsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		res, height, err := clientCtx.QueryWithData(
			fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryOracleScripts),
			bz,
		)
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

		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryRequests,
			vars[idTag],
		))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getRequestsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		res, height, err := clientCtx.QueryWithData(
			fmt.Sprintf("custom/%s/%s", oracletypes.QuerierRoute, oracletypes.QueryRequests),
			bz,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getRequestReportsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		params, ok := commonrest.СheckPaginationParams(w, r)
		if !ok {
			return
		}
		bz, err := clientCtx.LegacyAmino.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		res, height, err := clientCtx.QueryWithData(
			fmt.Sprintf(
				"custom/%s/%s/%s",
				oracletypes.QuerierRoute,
				oracletypes.QueryRequestReports,
				mux.Vars(r)[idTag],
			), bz,
		)
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

		oid, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("oid"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var callData []byte
		if r.FormValue("calldata") != "" {
			callData = []byte(r.FormValue("calldata"))
		}

		askCount, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("ask_count"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		minCount, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("min_count"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := oracleclientcommon.QuerySearchLatestRequest(
			oracletypes.QuerierRoute, clientCtx,
			oracletypes.NewQueryRequestSearchRequest(oid, callData, askCount, minCount),
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func getRequestsPricesHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPrices []oracletypes.QueryRequestPriceRequest

		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &requestPrices) {
			return
		}

		prices := make([]hookprice.Price, len(requestPrices))
		height := int64(0)
		for idx, requestPrice := range requestPrices {

			bin := clientCtx.LegacyAmino.MustMarshalJSON(oracletypes.NewQueryRequestPricesRequest(
				requestPrice.Symbol,
				requestPrice.MinCount,
				requestPrice.AskCount,
			))
			res, h, err := clientCtx.QueryWithData(fmt.Sprintf(
				"%s/%s",
				hookscommon.AppHook,
				oracletypes.QueryRequestPrices,
			), bin)
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

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, prices)
	}
}

// TODO: fix later
func getRequestsPriceSymbolsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		res, height, err := clientCtx.Query(fmt.Sprintf(
			"%s/%s",
			hookscommon.AppHook,
			oracletypes.QueryPriceSymbols,
		))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var symbols []string
		if err := clientCtx.LegacyAmino.UnmarshalBinaryBare(res, &symbols); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, symbols)
	}
}

func getMultiRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		_, _, limit, err := rest.ParseHTTPArgsWithLimit(r, 1)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		oid, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("oid"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var callData []byte
		if r.FormValue("calldata") != "" {
			callData = []byte(r.FormValue("calldata"))
		}

		askCount, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("ask_count"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		minCount, err := strconv.ParseInt(
			oracleclientcommon.ValueOrDefault(r.FormValue("min_count"), "0").(string), 10, 64,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		requestResponse, height, err := oracleclientcommon.QueryMultiSearchLatestRequest(
			oracletypes.QuerierRoute, clientCtx,
			oracletypes.NewQueryRequestSearchRequest(oid, callData, askCount, minCount), limit,
		)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if requestResponse == nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "specified request not found")
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, requestResponse)
	}
}

func getValidatorStatusHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryValidatorStatus,
			vars[validatorAddressTag],
		))
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
		bz, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryReporters,
			vars[validatorAddressTag],
		))
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

		bz, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryActiveValidators,
		))
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

		bz, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryDataProvidersPool,
		))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getDataProviderRewardHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		res, height, err := clientCtx.Query(fmt.Sprintf(
			"custom/%s/%s",
			oracletypes.QuerierRoute,
			oracletypes.QueryDataProviderReward,
		))
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
