package rest

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/x/oracle/client/common/proof"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/gorilla/mux"
)

const (
	idTag               = "idTag"
	limitTag            = "limitTag"
	offsetTag           = "offsetTag"
	dataHashTag         = "dataHashTag"
	validatorAddressTag = "validatorAddressTag"
)

func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryParams), getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryCounts), getCountsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryData, dataHashTag), getDataByHashHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryDataSources, idTag), getDataSourceByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}/{%s}", oracletypes.ModuleName, oracletypes.QueryDataSources, limitTag, offsetTag), getDataSourcesHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryOracleScripts, idTag), getOracleScriptByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}/{%s}", oracletypes.ModuleName, oracletypes.QueryOracleScripts, limitTag, offsetTag), getOracleScriptsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryRequests, idTag), getRequestByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}/{%s}", oracletypes.ModuleName, oracletypes.QueryRequests, limitTag, offsetTag), getRequestsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}/{%s}/{%s}", oracletypes.ModuleName, oracletypes.QueryRequestReports, idTag, limitTag, offsetTag), getRequestReportsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryRequestSearch), getRequestSearchHandler(clientCtx)).Methods("GET")
	// TODO: fix
	//rtr.HandleFunc(fmt.Sprintf("/%s/request_prices", oracletypes.ModuleName), getRequestsPricesHandler(clientCtx)).Methods("POST")
	//rtr.HandleFunc(fmt.Sprintf("/%s/price_symbols", oracletypes.ModuleName), getRequestsPriceSymbolsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryMultiRequestSearch), getMultiRequestSearchHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryValidatorStatus, validatorAddressTag), getValidatorStatusHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryReporters, validatorAddressTag), getReportersHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryDataProviderReward), getDataProviderRewardHandler(clientCtx)).Methods("GET")

	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryProof, proof.RequestIDTag), proof.GetProofHandlerFn(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryMultiProof), proof.GetMutiProofHandlerFn(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryRequestsCountProof), proof.GetRequestsCountProofHandlerFn(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryActiveValidators), getActiveValidatorsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryDataProvidersPool), dataProvidersPoolHandler(clientCtx)).Methods("GET")
	// TODO: add pending request REST API
}
