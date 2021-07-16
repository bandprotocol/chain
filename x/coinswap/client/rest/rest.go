package rest

import (
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc(fmt.Sprintf("/%s/params", coinswaptypes.ModuleName), getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/rate/{from}/{to}", coinswaptypes.ModuleName), getRateHandler(clientCtx)).Methods("GET")
}
