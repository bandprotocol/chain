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

	if !k.IsEligibleAccount(ctx, msg.Sender) {
		return nil, sdkerrors.Wrapf(minttypes.ErrAccountIsNotEligible, "account: %s", msg.Sender)
	}

	if k.LimitExceeded(ctx, msg.Amount) {
		return nil, sdkerrors.Wrapf(minttypes.ErrExceedsWithdrawalLimitPerTime, "amount: %s", msg.Amount.String())
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to parse receiver address %s", msg.Receiver)
	}

	if err := k.WithdrawCoinsFromTreasury(ctx, receiver, msg.Amount); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to mint %s coins to account %s", msg.Amount, msg.Receiver)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		minttypes.EventTypeWithdrawal,
		sdk.NewAttribute(minttypes.AttributeKeyWithdrawalAmount, msg.Amount.String()),
		sdk.NewAttribute(minttypes.AttributeKeyReceiver, msg.Receiver),
		sdk.NewAttribute(minttypes.AttributeKeySender, msg.Sender),
	))

	return &minttypes.MsgWithdrawCoinsToAccFromTreasuryResponse{}, nil
}
