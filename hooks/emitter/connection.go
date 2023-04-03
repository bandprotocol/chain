package emitter

import (
	"github.com/bandprotocol/chain/v2/hooks/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"
)

func (h *Hook) emitConnection(ctx sdk.Context, connectionId string) {
	conn, _ := h.connectionkeeper.GetConnection(ctx, connectionId)
	chainId := h.getChainIdFromClientId(ctx, conn.ClientId)
	h.Write("SET_CONNECTION", common.JsDict{
		"counterparty_chain_id":      chainId,
		"client_id":                  conn.GetClientID(),
		"connection_id":              connectionId,
		"counterparty_client_id":     conn.Counterparty.GetClientID(),
		"counterparty_connection_id": conn.Counterparty.ConnectionId,
	})
}

func (h *Hook) handleMsgConnectionOpenConfirm(ctx sdk.Context, msg *types.MsgConnectionOpenConfirm) {
	h.emitConnection(ctx, msg.ConnectionId)
}

func (h *Hook) handleMsgConnectionOpenAck(ctx sdk.Context, msg *types.MsgConnectionOpenAck) {
	h.emitConnection(ctx, msg.ConnectionId)
}
