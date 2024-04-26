package emitter

import (
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

func (h *Hook) emitRemoveFeed(signalID string) {
	h.Write("REMOVE_FEED", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitSetFeed(feed types.Feed) {
	h.Write("SET_FEED", common.JsDict{
		"signal_id":                      feed.SignalID,
		"power":                          feed.Power,
		"interval":                       feed.Interval,
		"last_interval_update_timestamp": feed.LastIntervalUpdateTimestamp,
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

// handleMsgSubmitSignals implements emitter handler for MsgSubmitSignals.
func (h *Hook) handleMsgSubmitSignals(
	ctx sdk.Context, txHash []byte, msg *types.MsgSubmitSignals, evMap common.EvMap, detail common.JsDict,
) {
	var involvedSignalIDs []string
	if signal_ids, ok := evMap[types.EventTypeSubmitSignals+"."+types.AttributeKeySignalID]; ok {
		involvedSignalIDs = append(involvedSignalIDs, signal_ids...)
	}
	if signal_ids, ok := evMap[types.EventTypeRemoveSignals+"."+types.AttributeKeySignalID]; ok {
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

	h.emitRemoveDelegatorSignals(msg.Delegator)
	for _, signal := range msg.Signals {
		h.emitSetDelegatorSignal(ctx, msg.Delegator, signal)
	}
}
