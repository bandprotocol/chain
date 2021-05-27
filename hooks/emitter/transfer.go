package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/bandprotocol/chain/hooks/common"
)

// handleMsgTransfer implements emitter handler for msgTransfer.
func (h *Hook) handleMsgTransfer(ctx sdk.Context, msg *types.MsgTransfer, evMap common.EvMap) {
	if events, ok := evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyData]; ok {
		packet := newPacket(
			ctx,
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcPort][0],
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcChannel][0],
			common.Atoui(evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySequence][0]),
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstPort][0],
			evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstChannel][0],
			false,
		)
		h.extractFungibleTokenPacket(ctx, []byte(events[0]), evMap, packet)
		h.Write("NEW_PACKET", packet)
	}
}
