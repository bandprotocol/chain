package emitter

import (
	"github.com/bandprotocol/chain/hooks/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcxfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

func (h *Hook) handleFungibleTokenPacket(
	ctx sdk.Context, txHash []byte, msg *types.MsgRecvPacket, packet common.JsDict, evMap common.EvMap, extra common.JsDict,
) bool {
	var data ibcxfertypes.FungibleTokenPacketData
	err := oracletypes.ModuleCdc.UnmarshalJSON(msg.Packet.Data, &data)
	if err == nil {
		packet["type"] = "fungible token"
		packet["data"] = common.JsDict{
			"denom":    data.Denom,
			"amount":   data.Amount,
			"sender":   data.Sender,
			"receiver": data.Receiver,
		}
		// TODO: patch this line when cosmos-sdk fix AttributeKeyAckSuccess value
		if evMap[ibcxfertypes.EventTypePacket+"."+ibcxfertypes.AttributeKeyAckSuccess][0] == "false" {
			packet["acknowledgement"] = common.JsDict{
				"success": true,
			}
		} else {
			packet["acknowledgement"] = common.JsDict{
				"success": false,
			}
		}
		h.Write("NEW_PACKET", packet)
		return true
	}
	return false
}

func (h *Hook) handleOracleRequestPacket(
	ctx sdk.Context, txHash []byte, msg *types.MsgRecvPacket, packet common.JsDict, evMap common.EvMap, extra common.JsDict,
) bool {
	var data oracletypes.OracleRequestPacketData
	err := oracletypes.ModuleCdc.UnmarshalJSON(msg.Packet.Data, &data)
	if err == nil {
		if events, ok := evMap[oracletypes.EventTypeRequest+"."+oracletypes.AttributeKeyID]; ok {
			id := oracletypes.RequestID(common.Atoi(events[0]))
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
			packet["type"] = "oracle request"
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
				"reason":  evMap[channeltypes.EventTypeWriteAck+"."+channeltypes.AttributeKeyAck][0],
			}
		}
		h.Write("NEW_PACKET", packet)
		return true
	}
	return false
}

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

	if ok := h.handleOracleRequestPacket(ctx, txHash, msg, packet, evMap, extra); ok {
		return
	}
	if ok := h.handleFungibleTokenPacket(ctx, txHash, msg, packet, evMap, extra); ok {
		return
	}
}

func (h *Hook) handleOracleResponsePacket(
	ctx sdk.Context, packet common.JsDict, evMap common.EvMap,
) bool {
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
		h.Write("NEW_PACKET", packet)
		return true
	}
	return false
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

	if ok := h.handleOracleResponsePacket(ctx, packet, evMap); ok {
		return
	}
}
