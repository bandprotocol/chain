package keeper

import (
	"context"
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ coinswaptypes.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func (m msgServer) Exchange(goCtx context.Context, msg *coinswaptypes.MsgExchange) (*coinswaptypes.MsgExchangeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	requesterAccAddr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid requester address")
	}

	err = m.ExchangeDenom(ctx, msg.From, msg.To, msg.Amount, requesterAccAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to exchange")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		coinswaptypes.EventTypeExchange,
		sdk.NewAttribute(coinswaptypes.AttributeKeyRequester, fmt.Sprintf("%s", msg.Requester)),
		sdk.NewAttribute(coinswaptypes.AttributeKeyAmount, fmt.Sprintf("%s", msg.Amount.String())),
		sdk.NewAttribute(coinswaptypes.AttributeKeyExchangeDenom, fmt.Sprintf("%s:%s", msg.From, msg.To)),
	))

	return &coinswaptypes.MsgExchangeResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the coinswap MsgServer interface for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) coinswaptypes.MsgServer {
	return &msgServer{Keeper: keeper}
}
