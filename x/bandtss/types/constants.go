package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetBandtssGrantMsgTypes get message types that can be granted.
func GetBandtssGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgHealthCheck{}),
	}
}
