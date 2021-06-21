package emitter

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
)

// handleMsgSend implements emitter handler for MsgSend.
func (h *Hook) handleMsgSend(msg *types.MsgSend) {
	h.AddAccountsInTx(msg.ToAddress)
}

// handleMsgMultiSend implements emitter handler for MsgMultiSend.
func (h *Hook) handleMsgMultiSend(msg *types.MsgMultiSend) {
	for _, output := range msg.Outputs {
		h.AddAccountsInTx(output.Address)
	}
}

func (h *Hook) handleEventTypeTransfer(evMap common.EvMap) {
	h.AddAccountsInBlock(evMap[types.EventTypeTransfer+"."+types.AttributeKeyRecipient][0])
}
