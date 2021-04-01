package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc("/coinswap/params", getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc("/coinswap/rate", getRateHandler(clientCtx)).Methods("GET")
}
