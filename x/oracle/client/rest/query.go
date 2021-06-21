package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
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
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s", types.QueryParams))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getCountsHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s", types.QueryCounts))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getDataByHashHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		res, _, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryData, vars[dataHashTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Disposition", "attachment;")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(res)
	}
}

func getDataSourceByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryDataSources, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getOracleScriptByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryOracleScripts, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

func getRequestByIDHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryRequests, vars[idTag]))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

// func getRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
// 		if !ok {
// 			return
// 		}
// 		bz, height, err := clientcmn.QuerySearchLatestRequest(
// 			clientCtx,
// 			r.FormValue("oid"), r.FormValue("calldata"), r.FormValue("ask_count"), r.FormValue("min_count"),
// 		)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		clientCtx = clientCtx.WithHeight(height)
// 		rest.PostProcessResponse(w, clientCtx, bz)
// 	}
// }

// func getRequestsPricesHandler(cliCtx context.CLIContext, route string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		decoder := json.NewDecoder(r.Body)
// 		var requestPrices RequestPrices
// 		err := decoder.Decode(&requestPrices)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 		prices := make([]price.Price, len(requestPrices.Symbols))
// 		height := int64(0)
// 		for idx, symbol := range requestPrices.Symbols {
// 			bz, h, err := cliCtx.Query(fmt.Sprintf("band/prices/%s/%d/%d", symbol, requestPrices.AskCount, requestPrices.MinCount))
// 			if h > height {
// 				height = h
// 			}
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 				return
// 			}
// 			var price price.Price
// 			err = cliCtx.Codec.UnmarshalBinaryBare(bz, &price)
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 				return
// 			}
// 			prices[idx] = price
// 		}
// 		bz, err := types.QueryOK(prices)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		clientcmn.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
// 	}
// }

// func getRequestsPriceSymbolsHandler(cliCtx context.CLIContext, route string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
// 		if !ok {
// 			return
// 		}

// 		bz, height, err := cliCtx.Query(fmt.Sprintf("band/price_symbols/%s/%s", r.FormValue("ask_count"), r.FormValue("min_count")))

// 		var symbols []string
// 		if err := cliCtx.Codec.UnmarshalBinaryBare(bz, &symbols); err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		bz, err = types.QueryOK(symbols)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}

// 		clientcmn.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
// 	}
// }

// func getMultiRequestSearchHandler(clientCtx client.Context) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
// 		if !ok {
// 			return
// 		}
// 		limit := 1
// 		if rawLimit := r.FormValue("limit"); rawLimit != "" {
// 			var err error
// 			limit, err = strconv.Atoi(rawLimit)
// 			if err != nil {
// 				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			}
// 		}
// 		bz, height, err := clientcmn.QueryMultiSearchLatestRequest(
// 			clientCtx, r.FormValue("oid"), r.FormValue("calldata"), r.FormValue("ask_count"), r.FormValue("min_count"), limit,
// 		)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
// 			return
// 		}
// 		clientCtx = clientCtx.WithHeight(height)
// 		rest.PostProcessResponse(w, clientCtx, bz)
// 	}
// }

func getValidatorStatusHandler(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}
		vars := mux.Vars(r)
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryValidatorStatus, vars[validatorAddressTag]))
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
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryReporters, vars[validatorAddressTag]))
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
		bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s", types.QueryActiveValidators))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, bz)
	}
}

// type requestDetail struct {
// 	ChainID    string           `json:"chain_id"`
// 	Validator  sdk.ValAddress   `json:"validator"`
// 	RequestID  types.RequestID  `json:"request_id,string"`
// 	ExternalID types.ExternalID `json:"external_id,string"`
// 	Reporter   string           `json:"reporter"`
// 	Signature  []byte           `json:"signature"`
// }

// func verifyRequest(cliCtx context.CLIContext, route string) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var detail requestDetail
// 		err := json.NewDecoder(r.Body).Decode(&detail)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 		reporterPubkey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeAccPub, detail.Reporter)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 		bz, height, err := clientcmn.VerifyRequest(
// 			route, cliCtx, detail.ChainID, detail.RequestID, detail.ExternalID,
// 			detail.Validator, reporterPubkey, detail.Signature,
// 		)
// 		if err != nil {
// 			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
// 			return
// 		}
// 		clientcmn.PostProcessQueryResponse(w, cliCtx.WithHeight(height), bz)
// 	}
// }
