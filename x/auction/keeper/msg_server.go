package keeper

import (
	"context"
	"fmt"
	auctiontypes "github.com/GeoDB-Limited/odin-core/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ auctiontypes.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func (m msgServer) BuyCoins(
	goCtx context.Context,
	msg *auctiontypes.MsgBuyCoins,
) (*auctiontypes.MsgBuyCoinsResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)
	auctionStatus := m.GetAuctionStatus(ctx)
	if !auctionStatus.Pending {
		return nil, sdkerrors.Wrapf(auctiontypes.ErrAuctionIsNotPending, "failed to buy coins")
	}

	requesterAccAddr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid requester address")
	}

	exchangeRates := m.GetExchangeRates(ctx)
	err = m.coinswapKeeper.ExchangeDenom(ctx, msg.From, msg.To, msg.Amount, requesterAccAddr, exchangeRates...)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to buy coins")
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		auctiontypes.EventTypeBuyCoins,
		sdk.NewAttribute(auctiontypes.AttributeKeyRequester, fmt.Sprintf("%s", msg.Requester)),
		sdk.NewAttribute(auctiontypes.AttributeKeyAmount, fmt.Sprintf("%s", msg.Amount.String())),
		sdk.NewAttribute(auctiontypes.AttributeKeyExchangeDenom, fmt.Sprintf("%s:%s", msg.From, msg.To)),
	))

	return &auctiontypes.MsgBuyCoinsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the auction MsgServer interface for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) auctiontypes.MsgServer {
	return &msgServer{Keeper: keeper}
}
