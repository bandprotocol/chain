package types

import (
	"github.com/GeoDB-Limited/odincore/chain/x/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgExchange(from types.Denom, to types.Denom, amt sdk.Coin, requester sdk.AccAddress) MsgExchange {
	return MsgExchange{
		From:      from,
		To:        to,
		Amount:    amt,
		Requester: requester,
	}
}
