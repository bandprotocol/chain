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
	rtr.HandleFunc(fmt.Sprintf("/%s/params", oracletypes.ModuleName), getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/counts", oracletypes.ModuleName), getCountsHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/data/{%s}", oracletypes.ModuleName, dataHashTag), getDataByHashHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/data_sources/{%s}", oracletypes.ModuleName, idTag), getDataSourceByIDHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/oracle_scripts/{%s}", oracletypes.ModuleName, idTag), getOracleScriptByIDHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/requests/{%s}", oracletypes.ModuleName, idTag), getRequestByIDHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/request_search", oracletypes.ModuleName), getRequestSearchHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/request_prices", oracletypes.ModuleName), getRequestsPricesHandler(clientCtx)).Methods("POST")
	//rtr.HandleFunc(fmt.Sprintf("/%s/price_symbols", oracletypes.ModuleName), getRequestsPriceSymbolsHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/multi_request_search", oracletypes.ModuleName), getMultiRequestSearchHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/validators/{%s}", oracletypes.ModuleName, validatorAddressTag), getValidatorStatusHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/reporters/{%s}", oracletypes.ModuleName, validatorAddressTag), getReportersHandler(clientCtx)).Methods("GET")
	////rtr.HandleFunc(fmt.Sprintf("/%s/proof/{%s}", oracletypes.ModuleName, proof.RequestIDTag), proof.GetProofHandlerFn(cliCtx, storeName)).Methods("GET")
	////rtr.HandleFunc(fmt.Sprintf("/%s/multi_proof", oracletypes.ModuleName), proof.GetMutiProofHandlerFn(cliCtx, storeName)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/active_validators", oracletypes.ModuleName), getActiveValidatorsHandler(clientCtx)).Methods("GET")
	//rtr.HandleFunc(fmt.Sprintf("/%s/verify_request", oracletypes.ModuleName), verifyRequest(clientCtx)).Methods("POST")
	//// Get the amount held in the oracle pool
	//rtr.HandleFunc(fmt.Sprintf("/%s/data_providers_pool", oracletypes.ModuleName), dataProvidersPoolHandler(clientCtx)).Methods("GET")
}
