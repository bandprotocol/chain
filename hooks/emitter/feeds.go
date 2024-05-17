package emitter

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

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

func (h *Hook) emitSetFeed(feed types.Feed) {
	h.Write("SET_FEED", common.JsDict{
		"signal_id":                      feed.SignalID,
		"power":                          feed.Power,
		"interval":                       feed.Interval,
		"last_interval_update_timestamp": feed.LastIntervalUpdateTimestamp * int64(math.Pow10(9)),
		"deviation_in_thousandth":        feed.DeviationInThousandth,
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

func (h *Hook) emitSetValidatorPrice(ctx sdk.Context, validator string, price types.SubmitPrice) {
	h.Write("SET_VALIDATOR_PRICE", common.JsDict{
		"validator": validator,
		"signal_id": price.SignalID,
		"price":     price.Price,
		"timestamp": ctx.BlockTime().UnixNano(),
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

func (h *Hook) emitSetPriceService(ctx sdk.Context, priceService types.PriceService) {
	h.Write("SET_PRICE_SERVICE", common.JsDict{
		"hash":      priceService.Hash,
		"version":   priceService.Version,
		"url":       priceService.Url,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

// handleMsgSubmitSignals implements emitter handler for MsgSubmitSignals.
func (h *Hook) handleMsgSubmitSignals(
	ctx sdk.Context, msg *types.MsgSubmitSignals, evMap common.EvMap,
) {
	h.emitRemoveDelegatorSignals(msg.Delegator)
	var involvedSignalIDs []string
	if signal_ids, ok := evMap[types.EventTypeUpdateFeed+"."+types.AttributeKeySignalID]; ok {
		involvedSignalIDs = append(involvedSignalIDs, signal_ids...)
	}
	if signal_ids, ok := evMap[types.EventTypeDeleteFeed+"."+types.AttributeKeySignalID]; ok {
		involvedSignalIDs = append(involvedSignalIDs, signal_ids...)
	}

	for _, signalID := range involvedSignalIDs {
		feed, err := h.feedsKeeper.GetFeed(ctx, signalID)
		if err != nil {
			h.emitRemoveFeed(signalID)
		} else {
			h.emitSetFeed(feed)
		}
	}

	for _, signal := range msg.Signals {
		h.emitSetDelegatorSignal(ctx, msg.Delegator, signal)
	}
}

// handleMsgSubmitPrices implements emitter handler for MsgSubmitPrices.
func (h *Hook) handleMsgSubmitPrices(
	ctx sdk.Context, msg *types.MsgSubmitPrices,
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

// handleMsgUpdatePriceService implements emitter handler for MsgUpdatePriceService.
func (h *Hook) handleMsgUpdatePriceService(
	ctx sdk.Context, msg *types.MsgUpdatePriceService,
) {
	h.emitSetPriceService(ctx, msg.PriceService)
}