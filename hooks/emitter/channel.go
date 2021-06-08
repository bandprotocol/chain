package emitter

import (
	"github.com/bandprotocol/chain/hooks/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcxfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

// TODO: update transfer acknowledgement for fungible token packet

func newPacket(ctx sdk.Context, srcPort string, srcChannel string, sequence uint64, dstPort string, dstChannel string, isIncoming bool) common.JsDict {
	return common.JsDict{
		"is_incoming":  isIncoming,
		"block_height": ctx.BlockHeight(),
		"src_channel":  srcChannel,
		"src_port":     srcPort,
		"sequence":     sequence,
		"dst_channel":  dstChannel,
		"dst_port":     dstPort,
	}
}

func (h *Hook) extractFungibleTokenPacket(
	ctx sdk.Context, dataOfPacket []byte, evMap common.EvMap, packet common.JsDict,
) bool {
	var data ibcxfertypes.FungibleTokenPacketData
	err := ibcxfertypes.ModuleCdc.UnmarshalJSON(dataOfPacket, &data)
	if err == nil {
		packet["type"] = "fungible token"
		packet["data"] = common.JsDict{
			"denom":    data.Denom,
			"amount":   data.Amount,
			"sender":   data.Sender,
			"receiver": data.Receiver,
		}
		if events, ok := evMap[ibcxfertypes.EventTypePacket+"."+ibcxfertypes.AttributeKeyAckSuccess]; ok {
			// TODO: patch this line when cosmos-sdk fix AttributeKeyAckSuccess value
			if events[0] == "false" {
				packet["acknowledgement"] = common.JsDict{
					"success": true,
				}
			} else {
				packet["acknowledgement"] = common.JsDict{
					"success": false,
				}
			}
		}
		return true
	}
	return false
}

func (h *Hook) extractOracleRequestPacket(
	ctx sdk.Context, txHash []byte, signer string, dataOfPacket []byte, evMap common.EvMap, msgJson common.JsDict, packet common.JsDict,
) bool {
	var data oracletypes.OracleRequestPacketData
	err := oracletypes.ModuleCdc.UnmarshalJSON(dataOfPacket, &data)
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
				"sender":           signer,
				"client_id":        data.ClientID,
				"resolve_status":   oracletypes.RESOLVE_STATUS_OPEN,
				"timestamp":        ctx.BlockTime().UnixNano(),
				"prepare_gas":      data.PrepareGas,
				"execute_gas":      data.ExecuteGas,
				"fee_limit":        data.FeeLimit.String(),
			})
			h.emitRawRequestAndValRequest(id, req)
			os := h.oracleKeeper.MustGetOracleScript(ctx, data.OracleScriptID)
			msgJson["id"] = id
			msgJson["name"] = os.Name
			msgJson["schema"] = os.Schema

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
		return true
	}
	return false
}

// handleMsgRequestData implements emitter handler for MsgRequestData.
func (h *Hook) handleMsgRecvPacket(
	ctx sdk.Context, txHash []byte, msg *types.MsgRecvPacket, evMap common.EvMap, msgJson common.JsDict,
) {
	packet := newPacket(
		ctx,
		msg.Packet.SourcePort,
		msg.Packet.SourceChannel,
		msg.Packet.Sequence,
		msg.Packet.DestinationPort,
		msg.Packet.DestinationChannel,
		true,
	)
	h.Write("NEW_PACKET", packet)
	if ok := h.extractOracleRequestPacket(ctx, txHash, msg.Signer, msg.Packet.Data, evMap, msgJson, packet); ok {
		return
	}
	if ok := h.extractFungibleTokenPacket(ctx, msg.Packet.Data, evMap, packet); ok {
		return
	}
}

func (h *Hook) extractOracleResponsePacket(
	ctx sdk.Context, packet common.JsDict, evMap common.EvMap,
) bool {
	var data oracletypes.OracleResponsePacketData
	err := oracletypes.ModuleCdc.UnmarshalJSON([]byte(evMap[types.EventTypeSendPacket+"."+types.AttributeKeyData][0]), &data)
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
		return true
	}
	return false
}

func (h *Hook) handleEventSendPacket(
	ctx sdk.Context, evMap common.EvMap,
) {
	packet := newPacket(
		ctx,
		evMap[types.EventTypeSendPacket+"."+types.AttributeKeySrcPort][0],
		evMap[types.EventTypeSendPacket+"."+types.AttributeKeySrcChannel][0],
		common.Atoui(evMap[types.EventTypeSendPacket+"."+types.AttributeKeySequence][0]),
		evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDstPort][0],
		evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDstChannel][0],
		false,
	)
	h.Write("NEW_PACKET", packet)
	if ok := h.extractOracleResponsePacket(ctx, packet, evMap); ok {
		return
	}
}
