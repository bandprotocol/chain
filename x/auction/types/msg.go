package types

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is the name of the auction module
const RouterKey = ModuleName

// Route returns the route of MsgBuyCoins - "additionalExchangeRates" (sdk.Msg interface).
func (msg MsgBuyCoins) Route() string { return RouterKey }

// Type returns the message type of MsgBuyCoins (sdk.Msg interface).
func (msg MsgBuyCoins) Type() string { return "buy_coins" }

// ValidateBasic checks whether the given MsgBuyCoins instance (sdk.Msg interface).
func (msg MsgBuyCoins) ValidateBasic() error {
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
		return sdkerrors.Wrapf(coinswaptypes.ErrInvalidExchangeDenom, "denominations: %s:%s", msg.From, msg.To)
	}
	if msg.From != msg.Amount.Denom {
		return sdkerrors.Wrapf(coinswaptypes.ErrExchangeDenomMissmatch, "denominations: %s:%s", msg.From, msg.To)
	}
	return nil
}

func (msg MsgBuyCoins) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgBuyCoins) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
