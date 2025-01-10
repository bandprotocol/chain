package emitter

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (h *Hook) emitSetTunnel(ctx sdk.Context, tunnelID uint64) {
	tunnel := h.tunnelKeeper.MustGetTunnel(ctx, tunnelID)
	latestSignalPrice, _ := h.tunnelKeeper.GetLatestPrices(ctx, tunnelID)
	h.Write("SET_TUNNEL", common.JsDict{
		"id":            tunnel.ID,
		"sequence":      tunnel.Sequence,
		"route_type":    tunnel.Route.TypeUrl,
		"route":         tunnel.Route.GetCachedValue(),
		"fee_payer":     tunnel.FeePayer,
		"total_deposit": tunnel.TotalDeposit.String(),
		"status":        tunnel.IsActive,
		"last_interval": latestSignalPrice.LastInterval * int64(time.Second),
		"creator":       tunnel.Creator,
		"created_at":    tunnel.CreatedAt * int64(time.Second),
	})
}

func (h *Hook) emitUpdateTunnelStatus(ctx sdk.Context, tunnelID uint64) {
	tunnel := h.tunnelKeeper.MustGetTunnel(ctx, tunnelID)
	h.Write("UPDATE_TUNNEL_STATUS", common.JsDict{
		"id":           tunnel.ID,
		"status":       tunnel.IsActive,
		"status_since": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetTunnelDeposit(ctx sdk.Context, tunnelID uint64, depositor string) {
	deposit, found := h.tunnelKeeper.GetDeposit(ctx, tunnelID, sdk.MustAccAddressFromBech32(depositor))
	if found {
		h.Write("SET_TUNNEL_DEPOSIT", common.JsDict{
			"tunnel_id":     deposit.TunnelID,
			"depositor":     deposit.Depositor,
			"total_deposit": deposit.Amount.String(),
		})
	} else {
		h.Write("REMOVE_TUNNEL_DEPOSIT", common.JsDict{
			"tunnel_id": tunnelID,
			"depositor": depositor,
		})
	}
}

func (h *Hook) emitSetTunnelHistoricalDeposit(
	ctx sdk.Context,
	txHash []byte,
	tunnelID uint64,
	depositor string,
	depositType int,
	amount sdk.Coins,
) {
	h.Write("SET_TUNNEL_HISTORICAL_DEPOSIT", common.JsDict{
		"tx_hash":      txHash,
		"tunnel_id":    tunnelID,
		"depositor":    depositor,
		"deposit_type": depositType,
		"amount":       amount.String(),
		"timestamp":    ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetTunnelHistoricalSignalDeviations(ctx sdk.Context, tunnelID uint64) {
	tunnel := h.tunnelKeeper.MustGetTunnel(ctx, tunnelID)

	h.Write("SET_TUNNEL_HISTORICAL_SIGNAL_DEVIATIONS", common.JsDict{
		"tunnel_id":         tunnel.ID,
		"created_at":        ctx.BlockTime().UnixNano(),
		"interval":          tunnel.Interval,
		"signal_deviations": tunnel.SignalDeviations,
	})
}

func (h *Hook) emitSetTunnelPacket(
	ctx sdk.Context,
	tunnelID uint64,
	sequence uint64,
	fees Fees,
) {
	packet, _ := h.tunnelKeeper.GetPacket(ctx, tunnelID, sequence)

	h.Write("SET_TUNNEL_PACKET", common.JsDict{
		"tunnel_id":    tunnelID,
		"sequence":     sequence,
		"receipt_type": packet.Receipt.TypeUrl,
		"receipt":      packet.Receipt.GetCachedValue(),
		"base_fee":     fees.BaseFee.String(),
		"route_fee":    fees.RouteFee.String(),
		"created_at":   packet.CreatedAt * int64(time.Second),
	})

	for _, sp := range packet.Prices {
		h.Write("SET_TUNNEL_PACKET_PRICE", common.JsDict{
			"tunnel_id": tunnelID,
			"sequence":  sequence,
			"signal_id": sp.SignalID,
			"status":    sp.Status,
			"price":     sp.Price,
			"timestamp": sp.Timestamp * int64(time.Second),
		})
	}
}

// handleTunnelMsgCreateTunnel implements emitter handler for MsgCreateTunnel.
func (h *Hook) handleTunnelMsgCreateTunnel(
	ctx sdk.Context,
	txHash []byte,
	msg *types.MsgCreateTunnel,
	evMap common.EvMap,
) {
	tunnelID := common.Atoui(evMap[types.EventTypeCreateTunnel+"."+types.AttributeKeyTunnelID][0])
	h.emitSetTunnel(ctx, tunnelID)
	h.emitSetTunnelHistoricalSignalDeviations(ctx, tunnelID)
	h.emitSetTunnelDeposit(ctx, tunnelID, msg.Creator)
	h.emitSetTunnelHistoricalDeposit(ctx, txHash, tunnelID, msg.Creator, 1, msg.InitialDeposit)
}

// handleTunnelMsgUpdateSignalsAndInterval implements emitter handler for MsgUpdateSignalsAndInterval.
func (h *Hook) handleTunnelMsgUpdateSignalsAndInterval(ctx sdk.Context, evMap common.EvMap) {
	tunnelID := common.Atoui(evMap[types.EventTypeUpdateSignalsAndInterval+"."+types.AttributeKeyTunnelID][0])
	h.emitSetTunnel(ctx, tunnelID)
	h.emitSetTunnelHistoricalSignalDeviations(ctx, tunnelID)
}

// handleTunnelMsgDepositToTunnel implements emitter handler for MsgDepositToTunnel.
func (h *Hook) handleTunnelMsgDepositToTunnel(ctx sdk.Context, txHash []byte, msg *types.MsgDepositToTunnel) {
	h.emitSetTunnel(ctx, msg.TunnelID)
	h.emitSetTunnelDeposit(ctx, msg.TunnelID, msg.Depositor)
	h.emitSetTunnelHistoricalDeposit(ctx, txHash, msg.TunnelID, msg.Depositor, 1, msg.Amount)
}

// handleTunnelMsgWithdrawFromTunnel implements emitter handler for MsgWithdrawFromTunnel.
func (h *Hook) handleTunnelMsgWithdrawFromTunnel(ctx sdk.Context, txHash []byte, msg *types.MsgWithdrawFromTunnel) {
	h.emitSetTunnel(ctx, msg.TunnelID)
	h.emitSetTunnelDeposit(ctx, msg.TunnelID, msg.Withdrawer)
	h.emitSetTunnelHistoricalDeposit(ctx, txHash, msg.TunnelID, msg.Withdrawer, 2, msg.Amount)
}

// handleTunnelMsgTriggerTunnel implements emitter handler for MsgTriggerTunnel.
func (h *Hook) handleTunnelMsgTriggerTunnel(
	ctx sdk.Context,
	msg *types.MsgTriggerTunnel,
	evMap common.EvMap,
	senderFeesMap map[string]Fees,
) {
	sequence := common.Atoui(evMap[types.EventTypeTriggerTunnel+"."+types.AttributeKeySequence][0])

	tunnel, _ := h.tunnelKeeper.GetTunnel(ctx, msg.TunnelID)

	h.emitSetTunnel(ctx, msg.TunnelID)
	h.emitSetTunnelPacket(ctx, msg.TunnelID, sequence, senderFeesMap[tunnel.FeePayer])
}

// handleTunnelEventTypeProducePacketSuccess implements emitter handler for EventTypeProducePacketSuccess.
func (h *Hook) handleTunnelEventTypeProducePacketSuccess(
	ctx sdk.Context,
	evMap common.EvMap,
	senderFeesMap map[string]Fees,
) {
	tunnelIDs := evMap[types.EventTypeProducePacketSuccess+"."+types.AttributeKeyTunnelID]
	sequences := evMap[types.EventTypeProducePacketSuccess+"."+types.AttributeKeySequence]

	for idx, tunnelID := range tunnelIDs {
		sequence := common.Atoui(sequences[idx])
		id := common.Atoui(tunnelID)

		tunnel := h.tunnelKeeper.MustGetTunnel(ctx, id)

		h.emitSetTunnel(ctx, id)
		h.emitSetTunnelPacket(ctx, id, sequence, senderFeesMap[tunnel.FeePayer])
	}
}

// handleTunnelEventTypeActivateTunnel implements emitter handler for EventTypeActivateTunnel.
func (h *Hook) handleTunnelEventTypeActivateTunnel(ctx sdk.Context, evMap common.EvMap) {
	tunnelIDs := evMap[types.EventTypeActivateTunnel+"."+types.AttributeKeyTunnelID]
	for _, tunnelID := range tunnelIDs {
		h.emitUpdateTunnelStatus(ctx, common.Atoui(tunnelID))
	}
}

// handleTunnelEventTypeDeactivateTunnel implements emitter handler for EventTypeDeactivateTunnel.
func (h *Hook) handleTunnelEventTypeDeactivateTunnel(ctx sdk.Context, evMap common.EvMap) {
	tunnelIDs := evMap[types.EventTypeDeactivateTunnel+"."+types.AttributeKeyTunnelID]
	for _, tunnelID := range tunnelIDs {
		h.emitUpdateTunnelStatus(ctx, common.Atoui(tunnelID))
	}
}
