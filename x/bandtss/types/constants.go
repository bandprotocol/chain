package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetGrantMsgTypes get message types that can be granted.
// NOTE: have to be a function, or else sdk cannot find msgTypeURL for granting.
func GetGrantMsgTypes() []string {
	return []string{sdk.MsgTypeURL(&MsgHealthCheck{})}
}
