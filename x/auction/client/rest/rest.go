package rest

import (
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

func RegisterRoutes(clientCtx client.Context, rtr *mux.Router) {
	rtr.HandleFunc(fmt.Sprintf("/%s/params", auctiontypes.ModuleName), getParamsHandler(clientCtx)).Methods("GET")
	rtr.HandleFunc(fmt.Sprintf("/%s/status", auctiontypes.ModuleName), getAuctionStatusHandler(clientCtx)).Methods("GET")
}
