package keeper

import (
	"context"
	"github.com/GeoDB-Limited/odin-core/x/mint/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the mint MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) WithdrawCoinsToAccFromTreasury(
	goCtx context.Context,
	msg *types.MsgWithdrawCoinsToAccFromTreasury,
) (*types.MsgWithdrawCoinsToAccFromTreasuryResponse, error) {
	return &types.MsgWithdrawCoinsToAccFromTreasuryResponse{}, nil
}
