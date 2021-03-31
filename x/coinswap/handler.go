package coinswap

import (
	"fmt"
	"github.com/GeoDB-Limited/odincore/chain/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case MsgExchange:
			return handleMsgExchange(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleMsgExchange(ctx sdk.Context, k Keeper, msg MsgExchange) (*sdk.Result, error) {
	validExchanges := k.GetValidExchangesParam(ctx, types.KeyValidExchanges)
	if !validExchanges.Contains(msg.From, msg.To) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "msg contains invalid denom exchange, from: %s, to: %s", msg.From, msg.To)
	}
	err := k.ExchangeDenom(ctx, msg.From, msg.To, msg.Amount, msg.Requester)
	if err != nil {
		return nil, err
	}
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeExchange,
		sdk.NewAttribute(types.AttributeKeyRequester, fmt.Sprintf("%s", msg.Requester.String())),
		sdk.NewAttribute(types.AttributeKeyAmount, fmt.Sprintf("%s", msg.Amount.String())),
		sdk.NewAttribute(types.AttributeKeyExchangeDenom, fmt.Sprintf("%s:%s", msg.From, msg.To)),
	))
	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
