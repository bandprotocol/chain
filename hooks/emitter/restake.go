package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/x/restake/types"
)

func (h *Hook) updateRestakeStake(ctx sdk.Context, stakerAddr string) {
	addr := sdk.MustAccAddressFromBech32(stakerAddr)
	stake := h.restakeKeeper.GetStake(ctx, addr)
	h.Write("SET_RESTAKE_STAKE", common.JsDict{
		"staker": stakerAddr,
		"coins":  stake.Coins.String(),
	})
}

// handleRestakeEventLockPower implements emitter handler for EventLockPower.
func (h *Hook) handleRestakeEventLockPower(_ sdk.Context, evMap common.EvMap) {
	h.Write("SET_RESTAKE_LOCK_POWER", common.JsDict{
		"staker": evMap[types.EventTypeLockPower+"."+types.AttributeKeyStaker][0],
		"key":    evMap[types.EventTypeLockPower+"."+types.AttributeKeyKey][0],
		"power":  evMap[types.EventTypeLockPower+"."+types.AttributeKeyPower][0],
	})
}

// handleRestakeEventStake implements emitter handler for EventStake.
func (h *Hook) handleRestakeEventStake(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeStake(ctx, evMap[types.EventTypeStake+"."+types.AttributeKeyStaker][0])
}

// handleRestakeEventUnstake implements emitter handler for EventUnstake.
func (h *Hook) handleRestakeEventUnstake(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeStake(ctx, evMap[types.EventTypeUnstake+"."+types.AttributeKeyStaker][0])
}
