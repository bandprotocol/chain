package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetBandtssGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&MsgHealthCheck{}),
	}
}
