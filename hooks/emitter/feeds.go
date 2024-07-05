package emitter

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (h *Hook) emitRemoveValidatorPrices(signalID string) {
	h.Write("REMOVE_VALIDATOR_PRICES", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitRemovePrice(signalID string) {
	h.emitRemoveValidatorPrices(signalID)
	h.Write("REMOVE_PRICE", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitRemoveFeed(signalID string) {
	h.emitRemovePrice(signalID)
	h.Write("REMOVE_FEED", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitSetSignalTotalPower(stp types.Signal) {
	h.Write("SET_SIGNAL_TOTAL_POWER", common.JsDict{
		"signal_id": stp.ID,
		"power":     stp.Power,
	})
}

func (h *Hook) emitRemoveSignalTotalPower(signalID string) {
	h.Write("REMOVE_SIGNAL_TOTAL_POWER", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitRemoveDelegatorSignals(delegator string) {
	h.Write("REMOVE_DELEGATOR_SIGNALS", common.JsDict{
		"delegator": delegator,
	})
}

func (h *Hook) emitSetDelegatorSignal(ctx sdk.Context, delegator string, signal types.Signal) {
	h.Write("SET_DELEGATOR_SIGNAL", common.JsDict{
		"delegator": delegator,
		"signal_id": signal.ID,
		"power":     signal.Power,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetValidatorPrice(ctx sdk.Context, validator string, price types.SignalPrice) {
	h.Write("SET_VALIDATOR_PRICE", common.JsDict{
		"validator":    validator,
		"price_status": price.PriceStatus.String(),
		"signal_id":    price.SignalID,
		"price":        price.Price,
		"timestamp":    ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetPrice(price types.Price) {
	h.Write("SET_PRICE", common.JsDict{
		"signal_id":    price.SignalID,
		"price_status": price.PriceStatus.String(),
		"price":        price.Price,
		"timestamp":    price.Timestamp * int64(math.Pow10(9)),
	})
}

func (h *Hook) emitSetReferenceSourceConfig(ctx sdk.Context, rsc types.ReferenceSourceConfig) {
	h.Write("SET_REFERENCE_SOURCE_CONFIG", common.JsDict{
		"ipfs_hash": rsc.IPFSHash,
		"version":   rsc.Version,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

// handleMsgSubmitSignals implements emitter handler for MsgSubmitSignals.
func (h *Hook) handleMsgSubmitSignals(
	ctx sdk.Context, msg *types.MsgSubmitSignals, evMap common.EvMap,
) {
	h.emitRemoveDelegatorSignals(msg.Delegator)

	signalIDs := evMap[types.EventTypeUpdateSignalTotalPower+"."+types.AttributeKeySignalID]

	for _, signalID := range signalIDs {
		stp, err := h.feedsKeeper.GetSignalTotalPower(ctx, signalID)
		if err == nil {
			h.emitSetSignalTotalPower(stp)
		}
	}

	for _, signal := range msg.Signals {
		h.emitSetDelegatorSignal(ctx, msg.Delegator, signal)
	}
}

// handleMsgSubmitSignalPrices implements emitter handler for MsgSubmitSignalPrices.
func (h *Hook) handleMsgSubmitSignalPrices(
	ctx sdk.Context, msg *types.MsgSubmitSignalPrices,
) {
	for _, price := range msg.Prices {
		h.emitSetValidatorPrice(ctx, msg.Validator, price)
	}
}

// handleEventUpdatePrice implements emitter handler for event UpdatePrice.
func (h *Hook) handleEventUpdatePrice(
	ctx sdk.Context, evMap common.EvMap,
) {
	if signal_ids, ok := evMap[types.EventTypeUpdatePrice+"."+types.AttributeKeySignalID]; ok {
		for _, signal_id := range signal_ids {
			price, err := h.feedsKeeper.GetPrice(ctx, signal_id)
			if err != nil {
				continue
			}
			h.emitSetPrice(price)
		}
	}
}

// handleMsgUpdateReferenceSourceConfig implements emitter handler for MsgUpdateReferenceSourceConfig.
func (h *Hook) handleMsgUpdateReferenceSourceConfig(
	ctx sdk.Context, msg *types.MsgUpdateReferenceSourceConfig,
) {
	h.emitSetReferenceSourceConfig(ctx, msg.ReferenceSourceConfig)
}
