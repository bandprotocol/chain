package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	uint64Len = 8
)

// GetGrantMsgTypes get message types that can be granted.
// NOTE: have to be a function, or else sdk cannot find msgTypeURL for granting.
func GetGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgSubmitDKGRound1{}),
		sdk.MsgTypeURL(&MsgSubmitDKGRound2{}),
		sdk.MsgTypeURL(&MsgComplain{}),
		sdk.MsgTypeURL(&MsgConfirm{}),
		sdk.MsgTypeURL(&MsgSubmitDEs{}),
		sdk.MsgTypeURL(&MsgSubmitSignature{}),
	}
}
