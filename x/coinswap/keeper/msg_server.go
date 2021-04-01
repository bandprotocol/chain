package keeper

import (
	"context"
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

func (m msgServer) Exchange(goCtx context.Context, msg *coinswaptypes.MsgExchange) (*coinswaptypes.MsgExchangeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validExchanges := m.GetValidExchangesParam(ctx, coinswaptypes.KeyValidExchanges)
	if !validExchanges.Contains(msg.From, msg.To) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "msg contains invalid denom exchange, from: %s, to: %s", msg.From, msg.To)
	}

	requesterAccAddr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid requester address")
	}

	err = m.ExchangeDenom(ctx, msg.From, msg.To, msg.Amount, requesterAccAddr)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		coinswaptypes.EventTypeExchange,
		sdk.NewAttribute(coinswaptypes.AttributeKeyRequester, fmt.Sprintf("%s", msg.Requester)),
		sdk.NewAttribute(coinswaptypes.AttributeKeyAmount, fmt.Sprintf("%s", msg.Amount.String())),
		sdk.NewAttribute(coinswaptypes.AttributeKeyExchangeDenom, fmt.Sprintf("%s:%s", msg.From, msg.To)),
	))
	return &coinswaptypes.MsgExchangeResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) coinswaptypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ coinswaptypes.MsgServer = msgServer{}
