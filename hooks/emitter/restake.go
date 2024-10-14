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

func (h *Hook) updateRestakeVault(ctx sdk.Context, key string) {
	vault, _ := h.restakeKeeper.GetVault(ctx, key)
	h.Write("SET_RESTAKE_VAULT", common.JsDict{
		"key":               vault.Key,
		"vault_address":     vault.VaultAddress,
		"is_active":         vault.IsActive,
		"rewards_per_power": vault.RewardsPerPower.String(),
		"total_power":       vault.TotalPower.String(),
		"remainders":        vault.Remainders.String(),
	})
}
func (h *Hook) updateRestakeLock(ctx sdk.Context, stakerAddr string, key string) {
	addr := sdk.MustAccAddressFromBech32(stakerAddr)
	lock, found := h.restakeKeeper.GetLock(ctx, addr, key)
	if !found {
		h.Write("REMOVE_RESTAKE_LOCK", common.JsDict{
			"key": key,
		})

		return
	}

	h.Write("SET_RESTAKE_LOCK", common.JsDict{
		"staker_address":   addr,
		"key":              key,
		"power":            lock.Power.String(),
		"pos_reward_debts": lock.PosRewardDebts.String(),
		"neg_reward_debts": lock.NegRewardDebts.String(),
	})
}

// handleRestakeEventClaimRewards implements emitter handler for EventClaimRewards.
func (h *Hook) handleRestakeEventClaimRewards(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeVault(ctx, evMap[types.EventTypeClaimRewards+"."+types.AttributeKeyKey][0])

}

// handleRestakeEventCreateVault implements emitter handler for EventCreateVault.
func (h *Hook) handleRestakeEventCreateVault(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeVault(ctx, evMap[types.EventTypeCreateVault+"."+types.AttributeKeyKey][0])
}

// handleRestakeEventAddRewards implements emitter handler for EventAddRewards.
func (h *Hook) handleRestakeEventAddRewards(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeVault(ctx, evMap[types.EventTypeAddRewards+"."+types.AttributeKeyKey][0])
}

// handleRestakeEventDeactivateVault implements emitter handler for EventDeactivateVault.
func (h *Hook) handleRestakeEventDeactivateVault(ctx sdk.Context, evMap common.EvMap) {
	h.updateRestakeVault(ctx, evMap[types.EventTypeDeactivateVault+"."+types.AttributeKeyKey][0])
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
