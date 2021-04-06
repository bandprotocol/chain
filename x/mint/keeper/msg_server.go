package keeper

import (
	"context"
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the mint MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) minttypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ minttypes.MsgServer = msgServer{}

func (k msgServer) WithdrawCoinsToAccFromTreasury(
	goCtx context.Context,
	msg *minttypes.MsgWithdrawCoinsToAccFromTreasury,
) (*minttypes.MsgWithdrawCoinsToAccFromTreasuryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	mintPool := k.GetMintPool(ctx)
	if !mintPool.IsEligibleAccount(msg.Sender) {
		return nil, sdkerrors.Wrapf(minttypes.ErrAccountIsNotEligible, "account: %s", msg.Sender)
	}

	if k.LimitExceeded(ctx, msg.Amount) {
		return nil, sdkerrors.Wrapf(minttypes.ErrExceedsWithdrawalLimitPerTime, "amount: %s", msg.Amount.String())
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to parse receiver address %s", msg.Receiver)
	}

	if msg.Amount.IsAllGT(mintPool.TreasuryPool) {
		return nil, sdkerrors.Wrapf(
			minttypes.ErrWithdrawalAmountExceedsModuleBalance,
			"withdrawal amount: %s exceeds %s module balance",
			msg.Amount.String(),
			minttypes.ModuleName,
		)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, receiver, msg.Amount); err != nil {
		return nil, sdkerrors.Wrapf(
			err,
			"failed to withdraw %s from %s module account",
			msg.Amount.String(),
			minttypes.ModuleName,
		)
	}

	mintPool.TreasuryPool = mintPool.TreasuryPool.Sub(msg.Amount)
	k.SetMintPool(ctx, mintPool)

	return &minttypes.MsgWithdrawCoinsToAccFromTreasuryResponse{}, nil
}
