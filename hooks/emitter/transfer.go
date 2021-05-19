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
		packet := h.getPacket(ctx, evMap, false)
		h.extractFungibleTokenPacket(ctx, []byte(events[0]), packet, evMap)
	}
}
