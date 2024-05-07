package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetGrantMsgTypes returns types for GrantMsg.
func GetGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgSubmitPrices{}),
	}
}
