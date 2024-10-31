package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SigningID is the type-safe unique identifier type for bandtss signing info.
type SigningID uint64

// GetGrantMsgTypes get message types that can be granted.
// NOTE: have to be a function, or else sdk cannot find msgTypeURL for granting.
func GetGrantMsgTypes() []string {
	return []string{sdk.MsgTypeURL(&MsgHeartbeat{})}
}
