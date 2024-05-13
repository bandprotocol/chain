package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	uint64Len = 8
)

var TSSGrantMsgTypes = []string{
	sdk.MsgTypeURL(&MsgSubmitDKGRound1{}),
	sdk.MsgTypeURL(&MsgSubmitDKGRound2{}),
	sdk.MsgTypeURL(&MsgComplain{}),
	sdk.MsgTypeURL(&MsgConfirm{}),
	sdk.MsgTypeURL(&MsgSubmitDEs{}),
	sdk.MsgTypeURL(&MsgSubmitSignature{}),
}
