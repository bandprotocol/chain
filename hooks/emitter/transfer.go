package emitter

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
)

// handleMsgTransfer implements emitter handler for msgTransfer.
func (h *Hook) handleMsgTransfer(
	ctx sdk.Context,
	txHash []byte,
	msg *types.MsgTransfer,
	evMap common.EvMap,
	detail common.JsDict,
) {
	if events, ok := evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDataHex]; ok {
		packet := newPacket(
			ctx,
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcPort][0],
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcChannel][0],
			common.Atoui(evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySequence][0]),
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstPort][0],
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstChannel][0],
			txHash,
		)

		event, _ := hex.DecodeString(events[0])
		h.extractFungibleTokenPacket(ctx, event, evMap, detail, packet)
		h.Write("NEW_OUTGOING_PACKET", packet)
	}
}
