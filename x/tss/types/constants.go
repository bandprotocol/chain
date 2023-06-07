package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var MsgGrants = []string{
	sdk.MsgTypeURL(&MsgCreateGroup{}),
	sdk.MsgTypeURL(&MsgSubmitDKGRound1{}),
	sdk.MsgTypeURL(&MsgSubmitDKGRound2{}),
	sdk.MsgTypeURL(&MsgComplain{}),
	sdk.MsgTypeURL(&MsgConfirm{}),
	sdk.MsgTypeURL(&MsgSubmitDEs{}),
	sdk.MsgTypeURL(&MsgSign{}),
}

const (
	AddrLen   = 20
	uint64Len = 8
)
