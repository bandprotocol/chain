package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

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
