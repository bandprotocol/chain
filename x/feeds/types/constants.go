package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgSubmitPrices{}),
	}
}
