package keeper

import (
	"context"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
)

type msgServer struct {
	Keeper
}

// todo
func (m msgServer) Exchange(ctx context.Context, exchange *coinswaptypes.MsgExchange) (*coinswaptypes.MsgExchangeResponse, error) {
	panic("implement me")
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) coinswaptypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ coinswaptypes.MsgServer = msgServer{}
