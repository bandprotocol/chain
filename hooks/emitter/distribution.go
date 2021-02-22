package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bandprotocol/chain/hooks/common"
)

func (h *Hook) getCurrentRewardAndCurrentRatio(ctx sdk.Context, addr sdk.ValAddress) (string, string) {
	currentReward := "0"
	currentRatio := "0"

	reward := h.distrKeeper.GetValidatorCurrentRewards(ctx, addr)
	latestReward := h.distrKeeper.GetValidatorHistoricalRewards(ctx, addr, reward.Period-1)

	if !reward.Rewards.IsZero() {
		currentReward = reward.Rewards[0].Amount.String()
	}
	if !latestReward.CumulativeRewardRatio.IsZero() {
		currentRatio = latestReward.CumulativeRewardRatio[0].Amount.String()
	}

	return currentReward, currentRatio
}

func (h *Hook) emitUpdateValidatorRewardAndAccumulatedCommission(ctx sdk.Context, addr sdk.ValAddress) {
	currentReward, currentRatio := h.getCurrentRewardAndCurrentRatio(ctx, addr)
	x := h.distrKeeper.GetValidatorAccumulatedCommission(ctx, addr)
	x.Commission.TruncateDecimal()
	accCommission, _ := h.distrKeeper.GetValidatorAccumulatedCommission(ctx, addr).Commission.TruncateDecimal()
	h.Write("UPDATE_VALIDATOR", common.JsDict{
		"operator_address":       addr.String(),
		"current_reward":         currentReward,
		"current_ratio":          currentRatio,
		"accumulated_commission": accCommission.String(),
	})
}

func (h *Hook) emitUpdateValidatorReward(ctx sdk.Context, addr sdk.ValAddress) {
	currentReward, currentRatio := h.getCurrentRewardAndCurrentRatio(ctx, addr)
	h.Write("UPDATE_VALIDATOR", common.JsDict{
		"operator_address": addr.String(),
		"current_reward":   currentReward,
		"current_ratio":    currentRatio,
	})
}

// handleMsgWithdrawDelegatorReward implements emitter handler for MsgWithdrawDelegatorReward.
func (h *Hook) handleMsgWithdrawDelegatorReward(
	ctx sdk.Context, msg *types.MsgWithdrawDelegatorReward, evMap common.EvMap, extra common.JsDict,
) {
	val, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	del, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	withdrawAddr := h.distrKeeper.GetDelegatorWithdrawAddr(ctx, del)
	h.AddAccountsInTx(withdrawAddr.String())
	h.emitUpdateValidatorReward(ctx, val)
	h.emitDelegationAfterWithdrawReward(ctx, val, withdrawAddr)
	extra["reward_amount"] = evMap[types.EventTypeWithdrawRewards+"."+sdk.AttributeKeyAmount][0]
}

// handleMsgSetWithdrawAddress implements emitter handler for MsgSetWithdrawAddress.
func (h *Hook) handleMsgSetWithdrawAddress(msg *types.MsgSetWithdrawAddress) {
	h.AddAccountsInTx(msg.WithdrawAddress)
}

// handleMsgWithdrawValidatorCommission implements emitter handler for MsgWithdrawValidatorCommission.
func (h *Hook) handleMsgWithdrawValidatorCommission(
	ctx sdk.Context, msg *types.MsgWithdrawValidatorCommission, evMap common.EvMap, extra common.JsDict,
) {
	val, _ := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	withdrawAddr := h.distrKeeper.GetDelegatorWithdrawAddr(ctx, val)
	h.AddAccountsInTx(withdrawAddr.String())
	h.emitUpdateValidatorRewardAndAccumulatedCommission(ctx, sdk.ValAddress(msg.ValidatorAddress))
	extra["commission_amount"] = evMap[types.EventTypeWithdrawCommission+"."+sdk.AttributeKeyAmount][0]
}
