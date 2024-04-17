package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// WithdrawRewards defines a method for creating a new validator
func (k msgServer) WithdrawRewards(
	goCtx context.Context,
	msg *types.MsgWithdrawRewards,
) (*types.MsgWithdrawRewardsResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	return &types.MsgWithdrawRewardsResponse{}, nil
}
