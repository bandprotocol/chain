package rest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/gorilla/mux"

	"github.com/bandprotocol/chain/v2/x/oracle/client/common/proof"
)

const (
	idTag               = "idTag"
	dataHashTag         = "dataHashTag"
	validatorAddressTag = "validatorAddressTag"
)

func RegisterHandlers(cliCtx client.Context, rtr *mux.Router) {
	r := rest.WithHTTPDeprecationHeaders(rtr)
	r.HandleFunc(fmt.Sprintf("/oracle/params"), getParamsHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/counts"), getCountsHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/data/{%s}", dataHashTag), getDataByHashHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/data_sources/{%s}", idTag), getDataSourceByIDHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/oracle_scripts/{%s}", idTag), getOracleScriptByIDHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/requests/{%s}", idTag), getRequestByIDHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/validators/{%s}", validatorAddressTag), getValidatorStatusHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/reporters/{%s}", validatorAddressTag), getReportersHandler(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/proof/{%s}", proof.RequestIDTag), proof.GetProofHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/multi_proof"), proof.GetMutiProofHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/requests_count_proof"), proof.GetRequestsCountProofHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/oracle/active_validators"), getActiveValidatorsHandler(cliCtx)).Methods("GET")
}
