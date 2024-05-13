package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func (h *Hook) emitSetBandtssStatus(member types.Member) {
	h.Write("SET_BANDTSS_STATUS", common.JsDict{
		"address":     member.Address,
		"is_active":   member.IsActive,
		"since":       member.Since.UnixNano(),
		"last_active": member.LastActive.UnixNano(),
	})
}

func (h *Hook) emitNewBandtssReplacement(replacement types.Replacement) {
	h.Write("NEW_BANDTSS_REPLACEMENT", common.JsDict{
		"tss_signing_id":   replacement.SigningID,
		"new_group_id":     replacement.NewGroupID,
		"new_pub_key":      parseBytes(replacement.NewPubKey),
		"current_group_id": replacement.CurrentGroupID,
		"current_pub_key":  parseBytes(replacement.CurrentPubKey),
		"exec_time":        replacement.ExecTime.UnixNano(),
		"status":           int(replacement.Status),
	})
}

func (h *Hook) emitUpdateBandtssReplacementStatus(status types.ReplacementStatus) {
	h.Write("UPDATE_BANDTSS_REPLACEMENT_STATUS", common.JsDict{
		"status": int(status),
	})
}

// handleUpdateBandtssStatus implements emitter handler for update bandtss status.
func (h *Hook) handleUpdateBandtssStatus(ctx sdk.Context, address sdk.AccAddress) {
	member, err := h.bandtssKeeper.GetMember(ctx, address)
	if err != nil {
		panic(err)
	}
	h.emitSetBandtssStatus(member)
}

// handleBandtssMsgActivate implements emitter handler for MsgActivate of bandtss.
func (h *Hook) handleBandtssMsgActivate(
	ctx sdk.Context, msg *types.MsgActivate,
) {
	acc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}

	h.handleUpdateBandtssStatus(ctx, acc)
}

// handleBandtssMsgHealthCheck implements emitter handler for MsgHealthCheck of bandtss.
func (h *Hook) handleBandtssMsgHealthCheck(
	ctx sdk.Context, msg *types.MsgHealthCheck,
) {
	acc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}

	h.handleUpdateBandtssStatus(ctx, acc)
}

// handleUpdateBandtssReplacementStatus implements emitter handler events related to replacements.
func (h *Hook) handleUpdateBandtssReplacementStatus(ctx sdk.Context) {
	r := h.bandtssKeeper.GetReplacement(ctx)
	if r.Status == types.REPLACEMENT_STATUS_SUCCESS {
		h.handleSetTSSGroup(ctx, r.CurrentGroupID)
	}

	h.emitUpdateBandtssReplacementStatus(r.Status)
}

// handleInitTSSReplacement implements emitter handler for init replacement event.
func (h *Hook) handleInitBandtssReplacement(ctx sdk.Context) {
	r := h.bandtssKeeper.GetReplacement(ctx)
	h.emitNewBandtssReplacement(r)
}
