package emitter

import (
	"encoding/hex"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
)

var (
	EventTypeCompleteUnbonding    = types.EventTypeCompleteUnbonding
	EventTypeCompleteRedelegation = types.EventTypeCompleteRedelegation
)

func (h *Hook) emitSetValidator(ctx sdk.Context, addr sdk.ValAddress) types.Validator {
	val, _ := h.stakingKeeper.GetValidator(ctx, addr)
	currentReward, currentRatio := h.getCurrentRewardAndCurrentRatio(ctx, addr)
	valAccComm, _ := h.distrKeeper.GetValidatorAccumulatedCommission(ctx, addr)
	accCommission, _ := valAccComm.Commission.TruncateDecimal()
	pub, _ := val.ConsPubKey()

	h.Write("SET_VALIDATOR", common.JsDict{
		"operator_address":       addr.String(),
		"delegator_address":      sdk.AccAddress(addr).String(),
		"consensus_address":      sdk.GetConsAddress(pub).String(),
		"consensus_pubkey":       hex.EncodeToString(pub.Bytes()),
		"moniker":                val.Description.Moniker,
		"identity":               val.Description.Identity,
		"website":                val.Description.Website,
		"details":                val.Description.Details,
		"commission_rate":        val.Commission.Rate.String(),
		"commission_max_rate":    val.Commission.MaxRate.String(),
		"commission_max_change":  val.Commission.MaxChangeRate.String(),
		"min_self_delegation":    val.MinSelfDelegation.String(),
		"tokens":                 val.Tokens.Uint64(),
		"jailed":                 val.Jailed,
		"active":                 val.Status == types.Bonded,
		"delegator_shares":       val.DelegatorShares.String(),
		"current_reward":         currentReward,
		"current_ratio":          currentRatio,
		"accumulated_commission": accCommission.String(),
		"last_update":            ctx.BlockTime().UnixNano(),
	})
	return val
}

func (h *Hook) emitUpdateValidator(ctx sdk.Context, addr sdk.ValAddress) (types.Validator, bool) {
	val, err := h.stakingKeeper.GetValidator(ctx, addr)
	if err != nil {
		h.Write("UPDATE_VALIDATOR", common.JsDict{
			"operator_address": addr.String(),
			"tokens":           0,
			"delegator_shares": "0",
			"current_reward":   "0",
			"current_ratio":    "0",
			"last_update":      ctx.BlockTime().UnixNano(),
		})
		return types.Validator{}, false
	}
	currentReward, currentRatio := h.getCurrentRewardAndCurrentRatio(ctx, addr)
	h.Write("UPDATE_VALIDATOR", common.JsDict{
		"operator_address": addr.String(),
		"tokens":           val.Tokens.Uint64(),
		"delegator_shares": val.DelegatorShares.String(),
		"current_reward":   currentReward,
		"current_ratio":    currentRatio,
		"jailed":           val.Jailed,
		"active":           val.Status == types.Bonded,
		"last_update":      ctx.BlockTime().UnixNano(),
	})
	return val, true
}

func (h *Hook) emitUpdateValidatorStatus(ctx sdk.Context, addr sdk.ValAddress) {
	status := h.oracleKeeper.GetValidatorStatus(ctx, addr)
	h.Write("UPDATE_VALIDATOR", common.JsDict{
		"operator_address": addr.String(),
		"status":           status.IsActive,
		"status_since":     status.Since.UnixNano(),
	})
}

func (h *Hook) emitDelegationAfterWithdrawReward(
	ctx sdk.Context,
	operatorAddress sdk.ValAddress,
	delegatorAddress sdk.AccAddress,
) {
	_, ratio := h.getCurrentRewardAndCurrentRatio(ctx, operatorAddress)
	h.Write("UPDATE_DELEGATION", common.JsDict{
		"delegator_address": delegatorAddress,
		"operator_address":  operatorAddress,
		"last_ratio":        ratio,
	})
}

func (h *Hook) emitDelegation(ctx sdk.Context, operatorAddress sdk.ValAddress, delegatorAddress sdk.AccAddress) {
	delegation, err := h.stakingKeeper.GetDelegation(ctx, delegatorAddress, operatorAddress)
	if err != nil {
		h.Write("REMOVE_DELEGATION", common.JsDict{
			"delegator_address": delegatorAddress,
			"operator_address":  operatorAddress,
		})
	} else {
		_, ratio := h.getCurrentRewardAndCurrentRatio(ctx, operatorAddress)
		h.Write("SET_DELEGATION", common.JsDict{
			"delegator_address": delegatorAddress,
			"operator_address":  operatorAddress,
			"shares":            delegation.Shares.String(),
			"last_ratio":        ratio,
		})
	}
}

// handleMsgCreateValidator implements emitter handler for MsgCreateValidator.
func (h *Hook) handleMsgCreateValidator(
	ctx sdk.Context, msg *types.MsgCreateValidator, detail common.JsDict,
) {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	val := h.emitSetValidator(ctx, valAddr)
	h.emitDelegation(ctx, valAddr, sdk.AccAddress(valAddr))
	detail["moniker"] = val.Description.Moniker
	detail["identity"] = val.Description.Identity
}

// handleMsgEditValidator implements emitter handler for MsgEditValidator.
func (h *Hook) handleMsgEditValidator(
	ctx sdk.Context, msg *types.MsgEditValidator, detail common.JsDict,
) {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	val := h.emitSetValidator(ctx, valAddr)
	detail["moniker"] = val.Description.Moniker
	detail["identity"] = val.Description.Identity
}

func (h *Hook) emitUpdateValidatorAndDelegation(
	ctx sdk.Context,
	operatorAddress sdk.ValAddress,
	delegatorAddress sdk.AccAddress,
) (types.Validator, bool) {
	val, found := h.emitUpdateValidator(ctx, operatorAddress)
	h.emitDelegation(ctx, operatorAddress, delegatorAddress)
	return val, found
}

// handleMsgDelegate implements emitter handler for MsgDelegate
func (h *Hook) handleMsgDelegate(
	ctx sdk.Context, msg *types.MsgDelegate, detail common.JsDict,
) {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	delAddr, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	val, _ := h.emitUpdateValidatorAndDelegation(ctx, valAddr, delAddr)
	detail["moniker"] = val.Description.Moniker
	detail["identity"] = val.Description.Identity
}

// handleMsgUndelegate implements emitter handler for MsgUndelegate
func (h *Hook) handleMsgUndelegate(
	ctx sdk.Context, msg *types.MsgUndelegate, evMap common.EvMap, detail common.JsDict,
) {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	delAddr, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	val, found := h.emitUpdateValidatorAndDelegation(ctx, valAddr, delAddr)
	h.emitUnbondingDelegation(ctx, msg, evMap)
	if found {
		detail["moniker"] = val.Description.Moniker
		detail["identity"] = val.Description.Identity
	} else {
		detail["moniker"] = "Removed"
		detail["identity"] = "Removed"
	}
}

func (h *Hook) emitUnbondingDelegation(ctx sdk.Context, msg *types.MsgUndelegate, evMap common.EvMap) {
	completeTime, _ := time.Parse(time.RFC3339, evMap[types.EventTypeUnbond+"."+types.AttributeKeyCompletionTime][0])
	coin, _ := sdk.ParseCoinNormalized(evMap[types.EventTypeUnbond+"."+sdk.AttributeKeyAmount][0])
	h.Write("NEW_UNBONDING_DELEGATION", common.JsDict{
		"delegator_address": msg.DelegatorAddress,
		"operator_address":  msg.ValidatorAddress,
		"creation_height":   ctx.BlockHeight(),
		"completion_time":   completeTime.UnixNano(),
		"amount":            coin.Amount.String(),
	})
}

// handleMsgBeginRedelegate implements emitter handler for MsgBeginRedelegate
func (h *Hook) handleMsgBeginRedelegate(
	ctx sdk.Context, msg *types.MsgBeginRedelegate, evMap common.EvMap, detail common.JsDict,
) {
	src, _ := sdk.ValAddressFromBech32(msg.ValidatorSrcAddress)
	dst, _ := sdk.ValAddressFromBech32(msg.ValidatorDstAddress)
	del, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	valSrc, found := h.emitUpdateValidatorAndDelegation(ctx, src, del)
	valDst, _ := h.emitUpdateValidatorAndDelegation(ctx, dst, del)
	h.emitUpdateRedelation(src, dst, del, evMap)
	if found {
		detail["val_src_moniker"] = valSrc.Description.Moniker
		detail["val_src_identity"] = valSrc.Description.Identity
	} else {
		detail["val_src_moniker"] = "Removed"
		detail["val_src_identity"] = "Removed"
	}
	detail["val_dst_moniker"] = valDst.Description.Moniker
	detail["val_dst_identity"] = valDst.Description.Identity
}

func (h *Hook) emitUpdateRedelation(
	operatorSrcAddress sdk.ValAddress,
	operatorDstAddress sdk.ValAddress,
	delegatorAddress sdk.AccAddress,
	evMap common.EvMap,
) {
	completeTime, _ := time.Parse(
		time.RFC3339,
		evMap[types.EventTypeRedelegate+"."+types.AttributeKeyCompletionTime][0],
	)
	coin, _ := sdk.ParseCoinNormalized(evMap[types.EventTypeRedelegate+"."+sdk.AttributeKeyAmount][0])
	h.Write("NEW_REDELEGATION", common.JsDict{
		"delegator_address":    delegatorAddress.String(),
		"operator_src_address": operatorSrcAddress.String(),
		"operator_dst_address": operatorDstAddress.String(),
		"completion_time":      completeTime.UnixNano(),
		"amount":               coin.Amount.String(),
	})
}

func (h *Hook) handleEventTypeCompleteUnbonding(ctx sdk.Context, evMap common.EvMap) {
	h.Write("REMOVE_UNBONDING", common.JsDict{"timestamp": ctx.BlockTime().UnixNano()})
	h.AddAccountsInBlock(evMap[types.EventTypeCompleteUnbonding+"."+types.AttributeKeyDelegator][0])
}

func (h *Hook) handEventTypeCompleteRedelegation(ctx sdk.Context) {
	h.Write("REMOVE_REDELEGATION", common.JsDict{"timestamp": ctx.BlockTime().UnixNano()})
}
