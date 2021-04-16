package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is the name of the coinswap module
const RouterKey = ModuleName

// Route returns the route of MsgExchange - "coinswap" (sdk.Msg interface).
func (msg MsgExchange) Route() string { return RouterKey }

// Type returns the message type of MsgExchange (sdk.Msg interface).
func (msg MsgExchange) Type() string { return "exchange" }

// ValidateBasic checks whether the given MsgExchange instance (sdk.Msg interface).
func (msg MsgExchange) ValidateBasic() error {
	requester, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(requester); err != nil {
		return sdkerrors.Wrapf(err, "requester: %s", requester)
	}
	if ok := msg.Amount.IsValid(); !ok {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "amount: %s", msg.Amount)
	}
	if msg.To == "" || msg.From == "" {
		return sdkerrors.Wrapf(ErrInvalidExchangeDenom, "denominations: %s:%s", msg.From, msg.To)
	}
	if msg.From != msg.Amount.Denom {
		return sdkerrors.Wrapf(ErrExchangeDenomMissmatch, "denominations: %s:%s", msg.From, msg.To)
	}
	return nil
}

func (msg MsgExchange) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgExchange) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
