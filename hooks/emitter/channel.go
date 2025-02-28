package emitter

import (
	"encoding/hex"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibcxfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (h *Hook) emitSetChannel(ctx sdk.Context, portId string, channelId string) {
	channel, _ := h.channelKeeper.GetChannel(ctx, portId, channelId)
	hop := channel.ConnectionHops[0]
	h.Write("SET_CHANNEL", common.JsDict{
		"connection_id":        hop,
		"port":                 portId,
		"counterparty_port":    channel.Counterparty.PortId,
		"channel":              channelId,
		"counterparty_channel": channel.Counterparty.ChannelId,
		"state":                channel.State,
		"order":                channel.Ordering,
		"last_update":          ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) handleMsgChannelOpenInit(ctx sdk.Context, msg *types.MsgChannelOpenInit, evMap common.EvMap) {
	h.emitSetChannel(ctx, msg.PortId, evMap[types.EventTypeChannelOpenInit+"."+types.AttributeKeyChannelID][0])
}

func (h *Hook) handleIcahostChannelOpenTry(ctx sdk.Context, msg *types.MsgChannelOpenTry) {
	counterpartyPortId := msg.Channel.Counterparty.PortId
	counterpartyAddress := strings.TrimPrefix(counterpartyPortId, "icacontroller-")
	connection := msg.Channel.ConnectionHops[0]
	acc, status := h.icahostKeeper.GetInterchainAccountAddress(ctx, connection, counterpartyPortId)

	h.AddAccountsInTx(acc)

	if status {
		h.Write("NEW_INTERCHAIN_ACCOUNT", common.JsDict{
			"address":              acc,
			"connection_id":        connection,
			"counterparty_port":    counterpartyPortId,
			"counterparty_address": counterpartyAddress,
		})
	}
}

func (h *Hook) handleMsgChannelOpenTry(ctx sdk.Context, msg *types.MsgChannelOpenTry, evMap common.EvMap) {
	switch msg.PortId {
	case "icahost":
		h.handleIcahostChannelOpenTry(ctx, msg)
	default:
		// No action needed for other cases yet
	}

	h.emitSetChannel(ctx, msg.PortId, evMap[types.EventTypeChannelOpenTry+"."+types.AttributeKeyChannelID][0])
}

func (h *Hook) handleMsgChannelOpenAck(ctx sdk.Context, msg *types.MsgChannelOpenAck) {
	h.emitSetChannel(ctx, msg.PortId, msg.ChannelId)
}

func (h *Hook) handleMsgChannelOpenConfirm(ctx sdk.Context, msg *types.MsgChannelOpenConfirm) {
	h.emitSetChannel(ctx, msg.PortId, msg.ChannelId)
}

func (h *Hook) handleMsgChannelCloseInit(ctx sdk.Context, msg *types.MsgChannelCloseInit) {
	h.emitSetChannel(ctx, msg.PortId, msg.ChannelId)
}

func (h *Hook) handleMsgChannelCloseConfirm(ctx sdk.Context, msg *types.MsgChannelCloseConfirm) {
	h.emitSetChannel(ctx, msg.PortId, msg.ChannelId)
}

func (h *Hook) handleMsgAcknowledgement(ctx sdk.Context, msg *types.MsgAcknowledgement, _ common.EvMap) {
	packet := common.JsDict{
		"src_channel": msg.Packet.SourceChannel,
		"src_port":    msg.Packet.SourcePort,
		"sequence":    msg.Packet.Sequence,
		"block_time":  ctx.BlockTime().UnixNano(),
	}

	var ack types.Acknowledgement
	if err := ibcxfertypes.ModuleCdc.UnmarshalJSON(msg.Acknowledgement, &ack); err == nil {
		switch resp := ack.Response.(type) {
		case *types.Acknowledgement_Result:
			packet["acknowledgement"] = common.JsDict{
				"status": "success",
			}
		case *types.Acknowledgement_Error:
			packet["acknowledgement"] = common.JsDict{
				"status": "failure",
				"reason": resp.Error,
			}
		}
	}

	var data ibcxfertypes.FungibleTokenPacketData
	err := ibcxfertypes.ModuleCdc.UnmarshalJSON(msg.Packet.GetData(), &data)
	if err == nil {
		// Update balance of sender (refund)
		h.AddAccountsInTx(data.Sender)
	}

	h.Write("UPDATE_OUTGOING_PACKET", packet)
}

func (h *Hook) handleMsgTimeout(ctx sdk.Context, msg *types.MsgTimeout) {
	packet := common.JsDict{
		"src_channel": msg.Packet.SourceChannel,
		"src_port":    msg.Packet.SourcePort,
		"sequence":    msg.Packet.Sequence,
		"block_time":  ctx.BlockTime().UnixNano(),
	}

	// We use acknowledgement column to track packet status
	packet["acknowledgement"] = common.JsDict{
		"status": "timeout",
	}

	// TODO: Handle other packet type
	var data ibcxfertypes.FungibleTokenPacketData
	err := ibcxfertypes.ModuleCdc.UnmarshalJSON(msg.Packet.GetData(), &data)
	if err == nil {
		// Update balance of sender (refund)
		h.AddAccountsInTx(data.Sender)
	}

	h.Write("UPDATE_OUTGOING_PACKET", packet)
}

func newPacket(
	ctx sdk.Context,
	srcPort string,
	srcChannel string,
	sequence uint64,
	dstPort string,
	dstChannel string,
	txHash []byte,
) common.JsDict {
	return common.JsDict{
		"block_height": ctx.BlockHeight(),
		"block_time":   ctx.BlockTime().UnixNano(),
		"src_channel":  srcChannel,
		"src_port":     srcPort,
		"sequence":     sequence,
		"dst_channel":  dstChannel,
		"dst_port":     dstPort,
		"hash":         txHash,
	}
}

func (h *Hook) extractFungibleTokenPacket(
	dataOfPacket []byte, evMap common.EvMap, detail common.JsDict, packet common.JsDict,
) bool {
	var data ibcxfertypes.FungibleTokenPacketData
	err := ibcxfertypes.ModuleCdc.UnmarshalJSON(dataOfPacket, &data)
	if err == nil {
		p := common.JsDict{
			"denom":    data.Denom,
			"amount":   data.Amount,
			"sender":   data.Sender,
			"receiver": data.Receiver,
		}
		detail["decoded_data"] = p
		detail["packet_type"] = "fungible_token"

		packet["type"] = "fungible_token"
		packet["data"] = p

		// Add Band account sender or receiver to account tx to update balance and related tx
		if _, err = sdk.AccAddressFromBech32(data.Sender); err == nil {
			h.AddAccountsInTx(data.Sender)
		}

		if _, err = sdk.AccAddressFromBech32(data.Receiver); err == nil {
			h.AddAccountsInTx(data.Receiver)
		}

		if events, ok := evMap[ibcxfertypes.EventTypePacket+"."+ibcxfertypes.AttributeKeyAckSuccess]; ok {
			if events[0] == "true" {
				packet["acknowledgement"] = common.JsDict{
					"status": "success",
				}
			} else {
				reason, _ := hex.DecodeString(evMap[types.EventTypeWriteAck+"."+types.AttributeKeyAckHex][0])
				packet["acknowledgement"] = common.JsDict{
					"status": "failure",
					"reason": string(reason),
				}
			}
		} else {
			packet["acknowledgement"] = common.JsDict{
				"status": "pending",
			}
		}
		return true
	}
	return false
}

func (h *Hook) extractOracleRequestPacket(
	ctx sdk.Context,
	txHash []byte,
	signer string,
	dataOfPacket []byte,
	evMap common.EvMap,
	detail common.JsDict,
	packet common.JsDict,
) bool {
	var data oracletypes.OracleRequestPacketData
	err := oracletypes.ModuleCdc.UnmarshalJSON(dataOfPacket, &data)
	if err == nil {
		if events, ok := evMap[oracletypes.EventTypeRequest+"."+oracletypes.AttributeKeyID]; ok {
			var prepareGasUsed uint64
			if eventRequestGasUsed, ok := evMap[oracletypes.EventTypeRequest+"."+oracletypes.AttributeKeyGasUsed]; ok {
				prepareGasUsed = ConvertToGas(common.Atoui(eventRequestGasUsed[0]))
			}

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
				"prepare_gas_used": prepareGasUsed,
				"execute_gas":      data.ExecuteGas,
				"execute_gas_used": uint64(0),
				"tss_encoder":      data.TSSEncoder,
				"fee_limit":        data.FeeLimit.String(),
				"total_fees":       evMap[oracletypes.EventTypeRequest+"."+oracletypes.AttributeKeyTotalFees][0],
				"is_ibc":           req.IBCChannel != nil,
			})
			h.emitRawRequestAndValRequest(ctx, id, req, evMap)
			os := h.oracleKeeper.MustGetOracleScript(ctx, data.OracleScriptID)
			data := common.JsDict{
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
			}
			detail["id"] = id
			detail["name"] = os.Name
			detail["schema"] = os.Schema
			detail["decoded_data"] = data
			detail["packet_type"] = "oracle_request"
			detail["skip"] = false

			packet["type"] = "oracle_request"
			packet["data"] = data
			packet["acknowledgement"] = common.JsDict{
				"status":     "success",
				"request_id": id,
			}
		} else {
			packet["type"] = "oracle_request"
			packet["data"] = common.JsDict{
				"oracle_script_id": data.OracleScriptID,
				"calldata":         parseBytes(data.Calldata),
				"ask_count":        data.AskCount,
				"min_count":        data.MinCount,
				"client_id":        data.ClientID,
				"prepare_gas":      data.PrepareGas,
				"execute_gas":      data.ExecuteGas,
				"fee_limit":        data.FeeLimit.String(),
			}
			reasons, ok := evMap[types.EventTypeWriteAck+"."+types.AttributeKeyAckHex]
			if !ok {
				detail["skip"] = true
				return false
			}

			reason, _ := hex.DecodeString(reasons[0])
			packet["acknowledgement"] = common.JsDict{
				"status": "failure",
				"reason": string(reason),
			}
		}
		return true
	}
	return false
}

func (h *Hook) extractInterchainAccountPacket(
	ctx sdk.Context,
	txHash []byte,
	dataOfPacket []byte,
	events []abci.Event,
	evMap common.EvMap,
	detail common.JsDict,
	packet common.JsDict,
) bool {
	var data icatypes.InterchainAccountPacketData
	err := icatypes.ModuleCdc.UnmarshalJSON(dataOfPacket, &data)
	if err == nil {
		var status string
		if events, ok := evMap[icatypes.EventTypePacket+"."+icatypes.AttributeKeyAckSuccess]; ok {
			if events[0] == "true" {
				status = "success"
				packet["acknowledgement"] = common.JsDict{
					"status": status,
				}
			} else {
				status = "failure"
				packet["acknowledgement"] = common.JsDict{
					"status": status,
					"reason": evMap[icatypes.EventTypePacket+"."+icatypes.AttributeKeyAckError][0],
				}
			}
		} else {
			return false
		}

		// extract and handle inner messages of packet
		var msgs []sdk.Msg
		var innerMessages []common.JsDict
		switch data.Type {
		case icatypes.EXECUTE_TX:
			msgs, _ = icatypes.DeserializeCosmosTx(h.cdc, data.Data, icatypes.EncodingProtobuf)
			for _, msg := range msgs {
				// add signers for this message into the transaction
				signers, _, err := h.cdc.GetMsgV1Signers(msg)
				if err != nil {
					continue
				}
				addrs := make([]string, len(signers))
				for idx, signer := range signers {
					addrs[idx] = sdk.AccAddress(signer).String()
				}
				h.AddAccountsInTx(addrs...)

				// decode message
				msgDetail := make(common.JsDict)
				DecodeMsg(msg, msgDetail)
				innerMessages = append(innerMessages, common.JsDict{
					"type": sdk.MsgTypeURL(msg),
					"msg":  msgDetail,
				})

				// call handler for this message if ack is success
				if status == "success" {
					h.handleMsg(ctx, txHash, msg, events, msgDetail)
				}
			}
		default:
			fmt.Print("got unspecified ica packet type")
		}

		packet["type"] = "interchain_account"
		packet["data"] = common.JsDict{
			"type": data.Type,
			"data": innerMessages,
			"memo": data.Memo,
		}

		detail["packet_type"] = "interchain_account"
		detail["decoded_data"] = common.JsDict{
			"type": data.Type,
			"data": innerMessages,
			"memo": data.Memo,
		}

		return true
	}
	return false
}

// handleMsgRequestData implements emitter handler for MsgRequestData.
func (h *Hook) handleMsgRecvPacket(
	ctx sdk.Context,
	txHash []byte,
	msg *types.MsgRecvPacket,
	events []abci.Event,
	evMap common.EvMap,
	detail common.JsDict,
) {
	packet := newPacket(
		ctx,
		msg.Packet.SourcePort,
		msg.Packet.SourceChannel,
		msg.Packet.Sequence,
		msg.Packet.DestinationPort,
		msg.Packet.DestinationChannel,
		txHash,
	)
	if _, ok := evMap[types.EventTypeWriteAck+"."+types.AttributeKeyDataHex]; ok {
		if ok := h.extractOracleRequestPacket(ctx, txHash, msg.Signer, msg.Packet.Data, evMap, detail, packet); ok {
			h.Write("NEW_INCOMING_PACKET", packet)
			return
		}
		if ok := h.extractFungibleTokenPacket(msg.Packet.Data, evMap, detail, packet); ok {
			h.Write("NEW_INCOMING_PACKET", packet)
			return
		}
		if ok := h.extractInterchainAccountPacket(ctx, txHash, msg.Packet.Data, events, evMap, detail, packet); ok {
			h.Write("NEW_INCOMING_PACKET", packet)
			return
		}
	}
}

func (h *Hook) extractOracleResponsePacket(
	ctx sdk.Context, packet common.JsDict, evMap common.EvMap,
) bool {
	var data oracletypes.OracleResponsePacketData

	eventData, _ := hex.DecodeString(evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDataHex][0])
	err := oracletypes.ModuleCdc.UnmarshalJSON(
		eventData,
		&data,
	)
	if err == nil {
		res := h.oracleKeeper.MustGetResult(ctx, data.RequestID)
		os := h.oracleKeeper.MustGetOracleScript(ctx, res.OracleScriptID)
		packet["type"] = "oracle response"
		packet["data"] = common.JsDict{
			"request_id":           data.RequestID,
			"oracle_script_id":     res.OracleScriptID,
			"oracle_script_name":   os.Name,
			"oracle_script_schema": os.Schema,
			"resolve_status":       data.ResolveStatus,
			"result":               parseBytes(data.Result),
		}
		return true
	}
	return false
}

func (h *Hook) extractTunnelPricesPacket(
	_ sdk.Context, packet common.JsDict, evMap common.EvMap,
) bool {
	var data tunneltypes.TunnelPricesPacketData

	eventData, _ := hex.DecodeString(evMap[types.EventTypeSendPacket+"."+types.AttributeKeyDataHex][0])
	err := tunneltypes.ModuleCdc.UnmarshalJSON(
		eventData,
		&data,
	)
	if err == nil {
		packet["type"] = "tunnel prices"
		packet["data"] = common.JsDict{
			"tunnel_id":  data.TunnelID,
			"sequence":   data.Sequence,
			"prices":     data.Prices,
			"created_at": data.CreatedAt,
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
		nil,
	)
	if ok := h.extractOracleResponsePacket(ctx, packet, evMap); ok {
		h.Write("NEW_OUTGOING_PACKET", packet)
		return
	} else if ok := h.extractTunnelPricesPacket(ctx, packet, evMap); ok {
		h.Write("NEW_OUTGOING_PACKET", packet)
		return
	}
}
