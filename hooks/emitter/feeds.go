package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

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

func (h *Hook) emitSetSignalPricesTx(ctx sdk.Context, txHash []byte, validator string, feeder string) {
	h.Write("SET_SIGNAL_PRICES_TX", common.JsDict{
		"tx_hash":   txHash,
		"validator": validator,
		"feeder":    feeder,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetValidatorPrices(ctx sdk.Context, validator string, prices []types.SignalPrice) {
	h.Write("SET_VALIDATOR_PRICES", common.JsDict{
		"validator": validator,
		"prices":    prices,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetPrices(ctx sdk.Context, prices []types.Price) {
	h.Write("SET_PRICES", common.JsDict{
		"prices":    prices,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetReferenceSourceConfig(ctx sdk.Context, rsc types.ReferenceSourceConfig) {
	h.Write("SET_REFERENCE_SOURCE_CONFIG", common.JsDict{
		"registry_ipfs_hash": rsc.RegistryIPFSHash,
		"registry_version":   rsc.RegistryVersion,
		"timestamp":          ctx.BlockTime().UnixNano(),
	})
}

// handleMsgSubmitSignals implements emitter handler for MsgSubmitSignals.
func (h *Hook) handleMsgSubmitSignals(
	ctx sdk.Context, msg *types.MsgSubmitSignals, evMap common.EvMap,
) {
	h.emitRemoveDelegatorSignals(msg.Delegator)

	updatedSignalIDs := evMap[types.EventTypeUpdateSignalTotalPower+"."+types.AttributeKeySignalID]
	deletedSignalIDs := evMap[types.EventTypeDeleteSignalTotalPower+"."+types.AttributeKeySignalID]

	for _, signalID := range updatedSignalIDs {
		stp, err := h.feedsKeeper.GetSignalTotalPower(ctx, signalID)
		if err != nil {
			h.emitRemoveSignalTotalPower(signalID)
		} else {
			h.emitSetSignalTotalPower(stp)
		}
	}

	for _, signalID := range deletedSignalIDs {
		h.emitRemoveSignalTotalPower(signalID)
	}

	for _, signal := range msg.Signals {
		h.emitSetDelegatorSignal(ctx, msg.Delegator, signal)
	}
}

// handleMsgSubmitSignalPrices implements emitter handler for MsgSubmitSignalPrices.
func (h *Hook) handleMsgSubmitSignalPrices(
	ctx sdk.Context,
	txHash []byte,
	msg *types.MsgSubmitSignalPrices,
	feeder string,
) {
	if feeder == "" {
		feeder = msg.Validator
	}

	h.emitSetSignalPricesTx(ctx, txHash, msg.Validator, feeder)
	h.emitSetValidatorPrices(ctx, msg.Validator, msg.Prices)
}

// handleEventUpdatePrice implements emitter handler for event UpdatePrice.
func (h *Hook) handleEventUpdatePrice(
	ctx sdk.Context,
) {
	prices := h.feedsKeeper.GetAllCurrentPrices(ctx)
	h.emitSetPrices(ctx, prices)
}

// handleMsgUpdateReferenceSourceConfig implements emitter handler for MsgUpdateReferenceSourceConfig.
func (h *Hook) handleMsgUpdateReferenceSourceConfig(
	ctx sdk.Context, msg *types.MsgUpdateReferenceSourceConfig,
) {
	h.emitSetReferenceSourceConfig(ctx, msg.ReferenceSourceConfig)
}
