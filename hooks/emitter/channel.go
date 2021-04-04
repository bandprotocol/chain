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
	packet := common.JsDict{
		"is_incoming":  true,
		"block_height": ctx.BlockHeight(),
		"src_channel":  msg.Packet.SourceChannel,
		"src_port":     msg.Packet.SourcePort,
		"sequence":     msg.Packet.Sequence,
		"dst_channel":  msg.Packet.DestinationChannel,
		"dst_port":     msg.Packet.DestinationPort,
	}

	// TODO: Check on other packet
	var data oracletypes.OracleRequestPacketData
	err := h.cdc.UnmarshalJSON(msg.Packet.Data, &data)
	if err == nil {
		// TODO: Find a way to check if cannot make a new request
		requestPassed := true
		if requestPassed {
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
				"prepare_gas":      data.PrepareGas,
				"execute_gas":      data.ExecuteGas,
				"fee_limit":        data.FeeLimit.String(),
			})
			h.emitRawRequestAndValRequest(id, req)
			os := h.oracleKeeper.MustGetOracleScript(ctx, data.OracleScriptID)
			extra["id"] = id
			extra["name"] = os.Name
			extra["schema"] = os.Schema

			packet["type"] = "oracle request"
			packet["data"] = common.JsDict{
				"oracle_script_id":     data.OracleScriptID,
				"oracle_script_name":   os.Name,
				"oracle_script_schema": os.Schema,
				"calldata":             parseBytes(data.Calldata),
				"ask_count":            data.AskCount,
				"min_count":            data.MinCount,
				"client_id":            data.ClientID,
				"prepare_gas":          data.PrepareGas,
				"execute_gas":          data.ExecuteGas,
				"fee_limit":            data.FeeLimit.String(),
				"request_key":          data.RequestKey,
			}
			packet["acknowledgement"] = common.JsDict{
				"success":    true,
				"request_id": id,
			}

		} else {
			packet["data"] = common.JsDict{
				"oracle_script_id": data.OracleScriptID,
				"calldata":         parseBytes(data.Calldata),
				"ask_count":        data.AskCount,
				"min_count":        data.MinCount,
				"client_id":        data.ClientID,
				"prepare_gas":      data.PrepareGas,
				"execute_gas":      data.ExecuteGas,
				"fee_limit":        data.FeeLimit.String(),
				"request_key":      data.RequestKey,
			}
			packet["acknowledgement"] = common.JsDict{
				"success": false,
				"reason":  "TODO",
			}
		}
	}

	h.Write("NEW_PACKET", packet)
}

func (h *Hook) handleEventSendPacket(
	ctx sdk.Context, evMap common.EvMap,
) {
	packet := common.JsDict{
		"is_incoming":  false,
		"block_height": ctx.BlockHeight(),
		"src_channel":  evMap[types.EventTypeSendPacket+"."+types.AttributeKeySrcChannel][0],
		"src_port":     evMap[types.EventTypeSendPacket+"."+types.AttributeKeySrcPort][0],
		"sequence":     common.Atoui(evMap[types.EventTypeSendPacket+"."+types.AttributeKeySequence][0]),
		"dst_channel":  evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDstChannel][0],
		"dst_port":     evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDstPort][0],
	}
	// TODO: Check on other packet
	var data oracletypes.OracleResponsePacketData
	err := h.cdc.UnmarshalJSON([]byte(evMap[types.EventTypeSendPacket+"."+types.AttributeKeyData][0]), &data)
	if err == nil {
		req := h.oracleKeeper.MustGetRequest(ctx, data.RequestID)
		os := h.oracleKeeper.MustGetOracleScript(ctx, req.OracleScriptID)
		packet["type"] = "oracle response"
		packet["data"] = common.JsDict{
			"request_id":           data.RequestID,
			"oracle_script_id":     req.OracleScriptID,
			"oracle_script_name":   os.Name,
			"oracle_script_schema": os.Schema,
			"resolve_status":       data.ResolveStatus,
			"result":               parseBytes(data.Result),
		}
	}

	h.Write("NEW_PACKET", packet)
}
