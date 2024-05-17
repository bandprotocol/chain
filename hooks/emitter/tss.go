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
		"status":          signing.Status,
		"created_height":  signing.CreatedHeight,
	})
}

func (h *Hook) emitUpdateTSSSigningSuccess(signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":        signing.ID,
		"status":    signing.Status,
		"signature": parseBytes(signing.Signature),
	})
}

func (h *Hook) emitUpdateTSSSigningFailed(reason string, signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": signing.Status,
		"reason": reason,
	})
}

func (h *Hook) emitUpdateTSSSigningStatus(signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": signing.Status,
	})
}

func (h *Hook) emitSetTSSGroup(group types.Group, dkgContext []byte) {
	h.Write("SET_TSS_GROUP", common.JsDict{
		"id":             group.ID,
		"size":           group.Size_,
		"threshold":      group.Threshold,
		"pub_key":        parseBytes(group.PubKey),
		"status":         group.Status,
		"dkg_context":    parseBytes(dkgContext),
		"module_owner":   group.ModuleOwner,
		"created_height": group.CreatedHeight,
	})
}

func (h *Hook) emitSetTSSMember(member types.Member) {
	h.Write("SET_TSS_MEMBER", common.JsDict{
		"id":           member.ID,
		"tss_group_id": member.GroupID,
		"address":      member.Address,
		"pub_key":      parseBytes(member.PubKey),
		"is_malicious": member.IsMalicious,
		"is_active":    member.IsActive,
	})
}

func (h *Hook) emitNewTSSAssignedMember(sid tss.SigningID, gid tss.GroupID, am types.AssignedMember) {
	h.Write("NEW_TSS_ASSIGNED_MEMBER", common.JsDict{
		"tss_signing_id": sid,
		"tss_group_id":   gid,
		"tss_member_id":  am.MemberID,
		"pub_d":          parseBytes(am.PubD),
		"pub_e":          parseBytes(am.PubE),
		"binding_factor": parseBytes(am.PubKey),
		"pub_nonce":      parseBytes(am.PubNonce),
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
			"status":          int(signing.Status),
			"created_height":  signing.CreatedHeight,
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
		h.emitSetTSSMember(m)
	}
}
