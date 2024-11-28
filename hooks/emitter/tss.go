package emitter

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (h *Hook) emitNewTSSSigning(
	signing types.Signing,
	contentType string,
	content string,
	originatorType string,
	originator string,
) {
	h.Write("NEW_TSS_SIGNING", common.JsDict{
		"id":                 signing.ID,
		"current_attempt":    signing.CurrentAttempt,
		"tss_group_id":       signing.GroupID,
		"encoded_originator": parseBytes(signing.Originator),
		"message":            parseBytes(signing.Message),
		"group_pub_key":      parseBytes(signing.GroupPubKey),
		"group_pub_nonce":    parseBytes(signing.GroupPubNonce),
		"status":             signing.Status,
		"created_height":     signing.CreatedHeight,
		"content_type":       contentType,
		"content":            content,
		"originator_type":    originatorType,
		"originator":         originator,
	})
}

func (h *Hook) emitUpdateTSSSigningAttempt(signing types.Signing) {
	h.Write("UPDATE_TSS_SIGNING", common.JsDict{
		"id":              signing.ID,
		"current_attempt": signing.CurrentAttempt,
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

func (h *Hook) emitNewTSSAssignedMember(
	sid tss.SigningID,
	attempt uint64,
	gid tss.GroupID,
	am types.AssignedMember,
) {
	h.Write("NEW_TSS_ASSIGNED_MEMBER", common.JsDict{
		"tss_signing_id":      sid,
		"tss_signing_attempt": attempt,
		"tss_group_id":        gid,
		"tss_member_id":       am.MemberID,
		"pub_d":               parseBytes(am.PubD),
		"pub_e":               parseBytes(am.PubE),
		"binding_factor":      parseBytes(am.PubKey),
		"pub_nonce":           parseBytes(am.PubNonce),
	})
}

func (h *Hook) emitUpdateTSSAssignedMember(
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
	sig tss.Signature,
	blockHeight int64,
) {
	h.Write("UPDATE_TSS_ASSIGNED_MEMBER", common.JsDict{
		"tss_signing_id":      signingID,
		"tss_signing_attempt": attempt,
		"tss_member_id":       memberID,
		"signature":           parseBytes(sig),
		"submitted_height":    blockHeight,
	})
}

// handleInitTSSModule implements emitter handler for initializing tss module.
func (h *Hook) handleInitTSSModule(ctx sdk.Context) {
	groups := h.tssKeeper.GetGroups(ctx)
	for _, group := range groups {
		h.emitSetTSSGroup(group, nil) // DKG data is already removed.
	}
}

// handleTSSEventCreateGroup implements emitter handler for CreateGroup event.
func (h *Hook) handleTSSEventCreateSigning(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeCreateSigning+"."+types.AttributeKeySigningID]
	contentTypes := evMap[types.EventTypeCreateSigning+"."+types.AttributeKeyContentType]
	contentInfos := evMap[types.EventTypeCreateSigning+"."+types.AttributeKeyContent]
	originatorTypes := evMap[types.EventTypeCreateSigning+"."+types.AttributeKeyOriginatorType]
	originatorInfos := evMap[types.EventTypeCreateSigning+"."+types.AttributeKeyOriginator]

	for i, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		h.emitNewTSSSigning(
			signing,
			contentTypes[i],
			contentInfos[i],
			originatorTypes[i],
			originatorInfos[i],
		)
	}
}

// handleTSSEventRequestSignature implements emitter handler for RequestSignature event.
func (h *Hook) handleTSSEventRequestSignature(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeRequestSignature+"."+types.AttributeKeySigningID]

	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))

		signing := h.tssKeeper.MustGetSigning(ctx, id)
		attempt := signing.CurrentAttempt
		attemptInfo := h.tssKeeper.MustGetSigningAttempt(ctx, id, attempt)

		h.emitUpdateTSSSigningAttempt(signing)

		for _, am := range attemptInfo.AssignedMembers {
			h.emitNewTSSAssignedMember(signing.ID, attempt, signing.GroupID, am)
		}
	}
}

// handleTSSEventSigningSuccess implements emitter handler for SigningSuccess event.
func (h *Hook) handleTSSEventSigningSuccess(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeSigningSuccess+"."+types.AttributeKeySigningID]
	for _, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		h.emitUpdateTSSSigningSuccess(signing)
	}
}

// handleTSSEventSigningFailed implements emitter handler for SigningSuccess event.
func (h *Hook) handleTSSEventSigningFailed(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeSigningFailed+"."+types.AttributeKeySigningID]
	errReasons := evMap[types.EventTypeSigningFailed+"."+types.AttributeKeyReason]
	for i, sid := range sids {
		id := tss.SigningID(common.Atoi(sid))
		signing := h.tssKeeper.MustGetSigning(ctx, id)

		errReason := "failed with some reason"
		if i < len(errReasons) {
			errReason = errReasons[i]
		}

		h.emitUpdateTSSSigningFailed(errReason, signing)
	}
}

// handleTSSSetGroup implements emitter handler events related to group.
func (h *Hook) handleTSSSetGroup(ctx sdk.Context, gid tss.GroupID) {
	group := h.tssKeeper.MustGetGroup(ctx, gid)
	dkgContext, err := h.tssKeeper.GetDKGContext(ctx, gid)
	if err != nil {
		dkgContext = []byte{}
	}

	h.emitSetTSSGroup(group, dkgContext)

	members := h.tssKeeper.MustGetMembers(ctx, gid)
	for _, m := range members {
		h.emitSetTSSMember(m)
	}
}

// handleTSSEventSubmitSignature implements emitter handler for SubmitSignature event.
func (h *Hook) handleTSSEventSubmitSignature(ctx sdk.Context, evMap common.EvMap) {
	sids := evMap[types.EventTypeSubmitSignature+"."+types.AttributeKeySigningID]
	attempts := evMap[types.EventTypeSubmitSignature+"."+types.AttributeKeyAttempt]
	memberIDs := evMap[types.EventTypeSubmitSignature+"."+types.AttributeKeyMemberID]
	sigs := evMap[types.EventTypeSubmitSignature+"."+types.AttributeKeySignature]

	if len(sids) != len(attempts) || len(sids) != len(memberIDs) || len(sids) != len(sigs) {
		panic("invalid event data")
	}

	for i, sid := range sids {
		signingID := tss.SigningID(common.Atoi(sid))
		attempt := uint64(common.Atoi(attempts[i]))
		memberID := tss.MemberID(common.Atoi(memberIDs[i]))

		bz, err := hex.DecodeString(sigs[i])
		if err != nil {
			panic("invalid signature")
		}
		signature := tss.Signature(bz)

		h.emitUpdateTSSAssignedMember(signingID, attempt, memberID, signature, ctx.BlockHeight())
	}
}
