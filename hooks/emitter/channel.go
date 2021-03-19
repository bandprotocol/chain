package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/hooks/common"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

// handleMsgRequestData implements emitter handler for MsgRequestData.
func (h *Hook) handleMsgRecvPacket(
	ctx sdk.Context, txHash []byte, msg *types.MsgRecvPacket, evMap common.EvMap, extra common.JsDict,
) {
	var data oracletypes.OracleRequestPacketData
	h.cdc.UnmarshalJSON(msg.Packet.Data, &data)
	id := oracletypes.RequestID(common.Atoi(evMap[oracletypes.EventTypeRequest+"."+oracletypes.AttributeKeyID][0]))
	req := h.oracleKeeper.MustGetRequest(ctx, id)
	h.Write("NEW_REQUEST", common.JsDict{
		"id":               id,
		"tx_hash":          txHash,
		"oracle_script_id": data.OracleScriptID,
		"calldata":         parseBytes(data.Calldata),
		"ask_count":        data.AskCount,
		"min_count":        data.MinCount,
		"sender":           msg.Signer,
		"client_id":        data.ClientID,
		"resolve_status":   oracletypes.RESOLVE_STATUS_OPEN,
		"timestamp":        ctx.BlockTime().UnixNano(),
	})
	h.emitRawRequestAndValRequest(id, req)
	os := h.oracleKeeper.MustGetOracleScript(ctx, data.OracleScriptID)
	extra["id"] = id
	extra["name"] = os.Name
	extra["schema"] = os.Schema
}
