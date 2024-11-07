package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (h *Hook) emitSetFeedsSignalTotalPower(stp types.Signal) {
	h.Write("SET_SIGNAL_TOTAL_POWER", common.JsDict{
		"signal_id": stp.ID,
		"power":     stp.Power,
	})
}

func (h *Hook) emitRemoveFeedsSignalTotalPower(signalID string) {
	h.Write("REMOVE_SIGNAL_TOTAL_POWER", common.JsDict{
		"signal_id": signalID,
	})
}

func (h *Hook) emitSetFeedsVoterSignal(ctx sdk.Context, voter string, signal types.Signal) {
	h.Write("SET_VOTER_SIGNAL", common.JsDict{
		"voter":     voter,
		"signal_id": signal.ID,
		"power":     signal.Power,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitRemoveFeedsVoterSignals(voter string) {
	h.Write("REMOVE_VOTER_SIGNALS", common.JsDict{
		"voter": voter,
	})
}

func (h *Hook) emitSetFeedsSignalPricesTx(ctx sdk.Context, txHash []byte, validator string, feeder string) {
	h.Write("SET_SIGNAL_PRICES_TX", common.JsDict{
		"tx_hash":   txHash,
		"validator": validator,
		"feeder":    feeder,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetFeedsValidatorPrices(ctx sdk.Context, validator string, prices []types.SignalPrice) {
	h.Write("SET_VALIDATOR_PRICES", common.JsDict{
		"validator": validator,
		"prices":    prices,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetFeedsPrices(ctx sdk.Context, prices []types.Price) {
	h.Write("SET_PRICES", common.JsDict{
		"prices":    prices,
		"timestamp": ctx.BlockTime().UnixNano(),
	})
}

func (h *Hook) emitSetFeedsReferenceSourceConfig(ctx sdk.Context, rsc types.ReferenceSourceConfig) {
	h.Write("SET_REFERENCE_SOURCE_CONFIG", common.JsDict{
		"registry_ipfs_hash": rsc.RegistryIPFSHash,
		"registry_version":   rsc.RegistryVersion,
		"timestamp":          ctx.BlockTime().UnixNano(),
	})
}

// handleMsgVote implements emitter handler for MsgVote.
func (h *Hook) handleFeedsMsgVote(
	ctx sdk.Context, msg *types.MsgVote, evMap common.EvMap,
) {
	h.emitRemoveFeedsVoterSignals(msg.Voter)

	updatedSignalIDs := evMap[types.EventTypeUpdateSignalTotalPower+"."+types.AttributeKeySignalID]
	deletedSignalIDs := evMap[types.EventTypeDeleteSignalTotalPower+"."+types.AttributeKeySignalID]

	for _, signalID := range updatedSignalIDs {
		stp, err := h.feedsKeeper.GetSignalTotalPower(ctx, signalID)
		if err != nil {
			h.emitRemoveFeedsSignalTotalPower(signalID)
		} else {
			h.emitSetFeedsSignalTotalPower(stp)
		}
	}

	for _, signalID := range deletedSignalIDs {
		h.emitRemoveFeedsSignalTotalPower(signalID)
	}

	for _, signal := range msg.Signals {
		h.emitSetFeedsVoterSignal(ctx, msg.Voter, signal)
	}
}

// handleMsgSubmitSignalPrices implements emitter handler for MsgSubmitSignalPrices.
func (h *Hook) handleFeedsMsgSubmitSignalPrices(
	ctx sdk.Context,
	txHash []byte,
	msg *types.MsgSubmitSignalPrices,
	feeder string,
) {
	if feeder == "" {
		feeder = msg.Validator
	}

	h.emitSetFeedsSignalPricesTx(ctx, txHash, msg.Validator, feeder)
	h.emitSetFeedsValidatorPrices(ctx, msg.Validator, msg.Prices)
}

// handleMsgUpdateReferenceSourceConfig implements emitter handler for MsgUpdateReferenceSourceConfig.
func (h *Hook) handleFeedsMsgUpdateReferenceSourceConfig(
	ctx sdk.Context, msg *types.MsgUpdateReferenceSourceConfig,
) {
	h.emitSetFeedsReferenceSourceConfig(ctx, msg.ReferenceSourceConfig)
}
