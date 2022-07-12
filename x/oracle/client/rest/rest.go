package rest

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/gorilla/mux"

	"github.com/bandprotocol/chain/v2/x/oracle/client/common/proof"
)

func RegisterHandlers(cliCtx client.Context, rtr *mux.Router) {
	// TODO: Move to grpc query
	r := rest.WithHTTPDeprecationHeaders(rtr)
	r.HandleFunc(fmt.Sprintf("/oracle/proof/{%s}", proof.RequestIDTag), proof.GetProofHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/oracle/multi_proof", proof.GetMutiProofHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/oracle/requests_count_proof", proof.GetRequestsCountProofHandlerFn(cliCtx)).Methods("GET")
}
