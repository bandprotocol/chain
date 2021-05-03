package rest

import (
	"fmt"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/gorilla/mux"
)

const (
	idTag               = "idTag"
	dataHashTag         = "dataHashTag"
	validatorAddressTag = "validatorAddressTag"
)

func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryParams), getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryCounts), getCountsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryData, dataHashTag), getDataByHashHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryDataSource, idTag), getDataSourceByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryDataSources), getDataSourcesHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryOracleScript, idTag), getOracleScriptByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryOracleScripts), getOracleScriptsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryRequest, idTag), getRequestByIDHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryRequests), getRequestsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryRequestReports, idTag), getRequestReportsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryRequestSearch), getRequestSearchHandler(clientCtx)).Methods("GET")
	// TODO: fix
	//rtr.HandleFunc(fmt.Sprintf("/%s/request_prices", oracletypes.ModuleName), getRequestsPricesHandler(clientCtx)).Methods("POST")
	//rtr.HandleFunc(fmt.Sprintf("/%s/price_symbols", oracletypes.ModuleName), getRequestsPriceSymbolsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryMultiRequestSearch), getMultiRequestSearchHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryValidatorStatus, validatorAddressTag), getValidatorStatusHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s/{%s}", oracletypes.ModuleName, oracletypes.QueryReporters, validatorAddressTag), getReportersHandler(clientCtx)).Methods("GET")
	// TODO: maybe remove ???
	//rtr.HandleFunc(fmt.Sprintf("/%s/proof/{%s}", oracletypes.ModuleName, proof.RequestIDTag), proof.GetProofHandlerFn(cliCtx, storeName)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/multi_proof", oracletypes.ModuleName), proof.GetMutiProofHandlerFn(cliCtx, storeName)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/verify_request", oracletypes.ModuleName), verifyRequest(clientCtx)).Methods("POST")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryActiveValidators), getActiveValidatorsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/%s", oracletypes.ModuleName, oracletypes.QueryDataProvidersPool), dataProvidersPoolHandler(clientCtx)).Methods("GET")
	// TODO: add pending request REST API
}
