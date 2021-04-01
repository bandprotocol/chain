package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgExchange(from string, to string, amt sdk.Coin, requester sdk.AccAddress) *MsgExchange {
	return &MsgExchange{
		From:      from,
		To:        to,
		Amount:    amt,
		Requester: requester.String(),
	}
}
