package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/GeoDB-Limited/odin-core/hooks/common"
)

// handleEventSlash implements emitter handler for Slashing event.
func (h *Hook) handleEventSlash(ctx sdk.Context, event common.EvMap) {
	if raw, ok := event[types.EventTypeSlash+"."+types.AttributeKeyJailed]; ok && len(raw) == 1 {
		consAddress, _ := sdk.ConsAddressFromBech32(raw[0])
		validator, _ := h.stakingKeeper.GetValidatorByConsAddr(ctx, consAddress)
		h.Write("UPDATE_VALIDATOR", common.JsDict{
			"operator_address": validator.OperatorAddress,
			"tokens":           validator.Tokens.Uint64(),
			"jailed":           validator.Jailed,
		})
	}
}

// handleMsgUnjail implements emitter handler for MsgUnjail.
func (h *Hook) handleMsgUnjail(
	ctx sdk.Context, msg *types.MsgUnjail,
) {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddr)
	validator, _ := h.stakingKeeper.GetValidator(ctx, valAddr)
	h.Write("UPDATE_VALIDATOR", common.JsDict{
		"operator_address": msg.ValidatorAddr,
		"jailed":           validator.Jailed,
	})
}
