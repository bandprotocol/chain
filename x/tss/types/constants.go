package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetTSSGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgSubmitDKGRound1{}),
		sdk.MsgTypeURL(&MsgSubmitDKGRound2{}),
		sdk.MsgTypeURL(&MsgComplain{}),
		sdk.MsgTypeURL(&MsgConfirm{}),
		sdk.MsgTypeURL(&MsgSubmitDEs{}),
		sdk.MsgTypeURL(&MsgSubmitSignature{}),
	}
}

const (
	uint64Len = 8
)
