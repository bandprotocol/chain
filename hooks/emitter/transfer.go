package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibcxfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/bandprotocol/chain/hooks/common"
)

// handleMsgTransfer implements emitter handler for msgTransfer.
func (h *Hook) handleMsgTransfer(ctx sdk.Context, msg *types.MsgTransfer, evMap common.EvMap) {
	if events, ok := evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyData]; ok {
		var data ibcxfertypes.FungibleTokenPacketData
		err := h.cdc.UnmarshalJSON([]byte(events[0]), &data)
		if err == nil {
			packet := common.JsDict{
				"is_incoming":  false,
				"block_height": ctx.BlockHeight(),
				"src_channel":  evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcChannel][0],
				"src_port":     evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySrcPort][0],
				"sequence":     common.Atoui(evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeySequence][0]),
				"dst_channel":  evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstChannel][0],
				"dst_port":     evMap[channeltypes.EventTypeSendPacket+"."+channeltypes.AttributeKeyDstPort][0],
			}
			packet["type"] = "fungible token"
			packet["data"] = common.JsDict{
				"denom":    data.Denom,
				"amount":   data.Amount,
				"sender":   data.Sender,
				"receiver": data.Receiver,
			}
			h.Write("NEW_PACKET", packet)
		}
	}
}
