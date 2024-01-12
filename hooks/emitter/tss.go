package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (h *Hook) emitNewTSSSigning(signing types.Signing) {
	h.Write("NEW_TSS_SIGNING", common.JsDict{
		"id":              signing.ID,
		"tss_group_id":    signing.GroupID,
		"group_pub_key":   parseBytes(signing.GroupPubKey),
		"msg":             parseBytes(signing.Message),
		"group_pub_nonce": parseBytes(signing.GroupPubNonce),
		"fee":             signing.Fee.String(),
		"status":          int(signing.Status),
		"created_height":  signing.CreatedHeight,
		"requester":       signing.Requester,
	})
}

func (h *Hook) emitUpdateTSSSigningSuccess(signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":        signing.ID,
		"status":    int(signing.Status),
		"signature": parseBytes(signing.Signature),
	})
}

func (h *Hook) emitUpdateTSSSigningFailed(reason string, signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": int(signing.Status),
		"reason": reason,
	})
}

func (h *Hook) emitUpdateTSSSigningStatus(signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": int(signing.Status),
	})
}

func (h *Hook) emitSetTSSStatus(status types.Status) {
	h.Write("SET_TSS_STATUS", common.JsDict{
		"address":     status.Address,
		"status":      int(status.Status),
		"since":       status.Since.UnixNano(),
		"last_active": status.LastActive.UnixNano(),
	})
}

func (h *Hook) emitSetTSSGroup(group types.Group, dkgContext []byte) {
	h.Write("SET_TSS_GROUP", common.JsDict{
		"id":                    group.ID,
		"size":                  group.Size_,
		"threshold":             group.Threshold,
		"dkg_context":           parseBytes(dkgContext),
		"pub_key":               parseBytes(group.PubKey),
		"status":                int(group.Status),
		"fee":                   group.Fee.String(),
		"latest_replacement_id": group.LatestReplacementID,
		"created_height":        group.CreatedHeight,
	})
}

func (h *Hook) emitSetTSSGroupMember(member types.Member) {
	h.Write("SET_TSS_GROUP_MEMBER", common.JsDict{
		"id":           member.ID,
		"tss_group_id": member.GroupID,
		"address":      member.Address,
		"pub_key":      parseBytes(member.PubKey),
		"is_malicious": member.IsMalicious,
	})
}

func (h *Hook) emitNewTSSAssignedMember(sid tss.SigningID, gid tss.GroupID, am types.AssignedMember) {
	h.Write("NEW_TSS_ASSIGNED_MEMBER", common.JsDict{
		"tss_signing_id":      sid,
		"tss_group_id":        gid,
		"tss_group_member_id": am.MemberID,
		"pub_d":               parseBytes(am.PubD),
		"pub_e":               parseBytes(am.PubE),
		"binding_factor":      parseBytes(am.PubKey),
		"pub_nonce":           parseBytes(am.PubNonce),
	})
}

func (h *Hook) emitNewTSSReplacement(replacement types.Replacement) {
	h.Write("NEW_TSS_REPLACEMENT", common.JsDict{
		"id":             replacement.ID,
		"tss_signing_id": replacement.SigningID,
		"from_group_id":  replacement.FromGroupID,
		"from_pub_key":   parseBytes(replacement.FromPubKey),
		"to_group_id":    replacement.ToGroupID,
		"to_pub_key":     parseBytes(replacement.ToPubKey),
		"exec_time":      replacement.ExecTime.UnixNano(),
		"status":         int(replacement.Status),
	})
}

func (h *Hook) emitUpdateTSSReplacementStatus(ctx sdk.Context, id uint64, status types.ReplacementStatus) {
	h.Write("UPDATE_TSS_REPLACEMENT_STATUS", common.JsDict{
		"id":     id,
		"status": int(status),
	})
}

// handleInitTSSModule implements emitter handler for initializing tss module.
func (h *Hook) handleInitTSSModule(ctx sdk.Context) {
	for _, signing := range h.tssKeeper.GetSignings(ctx) {
		h.Write("NEW_TSS_SIGNING", common.JsDict{
			"id":              signing.ID,
			"tss_group_id":    signing.GroupID,
			"group_pub_key":   parseBytes(signing.GroupPubKey),
			"msg":             parseBytes(signing.Message),
			"group_pub_nonce": parseBytes(signing.GroupPubNonce),
			"signature":       parseBytes(signing.Signature),
			"fee":             signing.Fee.String(),
			"status":          int(signing.Status),
			"created_height":  signing.CreatedHeight,
			"requester":       signing.Requester,
		})
	}
}

// handleEventRequestSignature implements emitter handler for RequestSignature event.
func (h *Hook) handleEventRequestSignature(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeRequestSignature+"."+types.AttributeKeySigningID]
	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		h.emitNewTSSSigning(signing)

		for _, am := range signing.AssignedMembers {
			h.emitNewTSSAssignedMember(signing.ID, signing.GroupID, am)
		}
	}
}

// handleEventSigningSuccess implements emitter handler for SigningSuccess event.
func (h *Hook) handleEventSigningSuccess(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeSigningSuccess+"."+types.AttributeKeySigningID]
	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		h.emitUpdateTSSSigningSuccess(signing)
	}
}

// handleEventSigningFailed implements emitter handler for SigningSuccess event.
func (h *Hook) handleEventSigningFailed(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeSigningFailed+"."+types.AttributeKeySigningID]
	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		if reason, ok := evMap[types.EventTypeSigningFailed+"."+types.AttributeKeyReason]; ok {
			h.emitUpdateTSSSigningFailed(
				reason[0],
				signing,
			)
		} else {
			h.emitUpdateTSSSigningFailed(
				"failed with on reason",
				signing,
			)
		}
	}
}

// handleEventExpiredSigning implements emitter handler for ExpiredSigning event.
func (h *Hook) handleEventExpiredSigning(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeExpiredSigning+"."+types.AttributeKeySigningID]
	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		h.emitUpdateTSSSigningStatus(signing)
	}
}

// handleUpdateTSSStatus implements emitter handler for update tss status.
func (h *Hook) handleUpdateTSSStatus(ctx sdk.Context, address sdk.AccAddress) {
	status := h.tssKeeper.GetStatus(ctx, address)
	h.emitSetTSSStatus(status)
}

// handleSetTSSGroup implements emitter handler events related to group.
func (h *Hook) handleSetTSSGroup(ctx sdk.Context, gid tss.GroupID) {
	group := h.tssKeeper.MustGetGroup(ctx, gid)
	dkgContext, err := h.tssKeeper.GetDKGContext(ctx, gid)
	if err != nil {
		panic(err)
	}

	h.emitSetTSSGroup(group, dkgContext)

	members := h.tssKeeper.MustGetMembers(ctx, gid)
	for _, m := range members {
		h.emitSetTSSGroupMember(m)
	}
}

// handleInitTSSReplacement implements emitter handler for init replacement event.
func (h *Hook) handleInitTSSReplacement(ctx sdk.Context, evMap common.EvMap) {
	rids := evMap[types.EventTypeReplacement+"."+types.AttributeKeyReplacementID]
	for _, rid := range rids {
		id := uint64(common.Atoi(rid))
		r, err := h.tssKeeper.GetReplacement(ctx, id)
		if err != nil {
			panic(err)
		}

		h.emitNewTSSReplacement(r)
	}
}

// handleUpdateTSSReplacementStatus implements emitter handler events related to replacements.
func (h *Hook) handleUpdateTSSReplacementStatus(ctx sdk.Context, rid uint64) {
	r, err := h.tssKeeper.GetReplacement(ctx, rid)
	if err != nil {
		panic(err)
	}
	if r.Status == types.REPLACEMENT_STATUS_SUCCESS {
		h.handleSetTSSGroup(ctx, r.ToGroupID)
	}

	h.emitUpdateTSSReplacementStatus(ctx, rid, r.Status)
}

// handleTSSMsgActivate implements emitter handler for MsgActivate of TSS.
func (h *Hook) handleTSSMsgActivate(
	ctx sdk.Context, msg *types.MsgActivate,
) {
	acc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}

	h.handleUpdateTSSStatus(ctx, acc)
}

// handleTSSMsgHealthCheck implements emitter handler for MsgHealthCheck of TSS.
func (h *Hook) handleTSSMsgHealthCheck(
	ctx sdk.Context, msg *types.MsgHealthCheck,
) {
	acc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}

	h.handleUpdateTSSStatus(ctx, acc)
}

// handleTSSMsgSubmitDEs implements emitter handler for MsgSubmitDEs of TSS.
func (h *Hook) handleTSSMsgSubmitDEs(
	ctx sdk.Context, msg *types.MsgSubmitDEs,
) {
	acc, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}

	h.handleUpdateTSSStatus(ctx, acc)
}
