package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewMsgBuyCoins(from string, to string, amt sdk.Coin, requester sdk.AccAddress) *MsgBuyCoins {
	return &MsgBuyCoins{
		From:      from,
		To:        to,
		Amount:    amt,
		Requester: requester.String(),
	}
}
