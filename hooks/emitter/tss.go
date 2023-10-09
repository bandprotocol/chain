package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (h *Hook) emitNewSigning(signing tsstypes.Signing) {
	h.Write("NEW_SIGNING", common.JsDict{
		"id":              signing.ID,
		"group_id":        signing.GroupID,
		"group_pub_key":   parseBytes(signing.GroupPubKey),
		"msg":             parseBytes(signing.Message),
		"group_pub_nonce": parseBytes(signing.GroupPubNonce),
		"fee":             signing.Fee.String(),
		"status":          int(signing.Status),
		"created_height":  signing.CreatedHeight,
		"requester":       signing.Requester,
	})
}

func (h *Hook) emitUpdateSigningSuccess(signing tsstypes.Signing) {
	h.Write("UPDATE_SIGNING", common.JsDict{
		"id":        signing.ID,
		"status":    int(signing.Status),
		"signature": parseBytes(signing.Signature),
	})
}

func (h *Hook) emitUpdateSigningFailed(reason string, signing tsstypes.Signing) {
	h.Write("UPDATE_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": int(signing.Status),
		"reason": reason,
	})
}

// future use
func (h *Hook) emitUpdateSigningExpired(signing tsstypes.Signing) {
	h.Write("UPDATE_SIGNING", common.JsDict{
		"id":     signing.ID,
		"status": int(signing.Status),
	})
}

func (h *Hook) emitSetTSSAccountStatus(status tsstypes.Status) {
	h.Write("SET_TSS_ACCOUNT_STATUS", common.JsDict{
		"address":     status.Address,
		"status":      int(status.Status),
		"since":       status.Since.Unix(),
		"last_active": status.LastActive.Unix(),
	})
}

func (h *Hook) emitSetGroup(group tsstypes.Group, dkgContext []byte) {
	h.Write("SET_GROUP", common.JsDict{
		"id":                    group.GroupID,
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

func (h *Hook) emitSetMember(member tsstypes.Member) {
	h.Write("SET_MEMBER", common.JsDict{
		"id":           member.ID,
		"group_id":     member.GroupID,
		"address":      member.Address,
		"pub_key":      parseBytes(member.PubKey),
		"is_malicious": member.IsMalicious,
	})
}

func (h *Hook) emitNewAssignedMember(sid tss.SigningID, gid tss.GroupID, am tsstypes.AssignedMember) {
	h.Write("NEW_ASSIGNED_MEMBER", common.JsDict{
		"signing_id":     sid,
		"member_id":      am.MemberID,
		"group_id":       gid,
		"pub_d":          parseBytes(am.PubD),
		"pub_e":          parseBytes(am.PubE),
		"binding_factor": parseBytes(am.PubKey),
		"pub_nonce":      parseBytes(am.PubNonce),
	})
}

// handleInitTssModule implements emitter handler for initializing tss module.
func (h *Hook) handleInitTssModule(ctx sdk.Context) {
	for _, signing := range h.tssKeeper.GetAllReplacementSigning(ctx) {
		h.Write("NEW_SIGNING", common.JsDict{
			"id":              signing.ID,
			"group_id":        signing.GroupID,
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
	id := tss.SigningID(common.Atoi(evMap[tsstypes.EventTypeRequestSignature+"."+types.AttributeKeySigningID][0]))
	signing := h.tssKeeper.MustGetSigning(ctx, id)

	for _, am := range signing.AssignedMembers {
		h.emitNewAssignedMember(signing.ID, signing.GroupID, am)
	}

	h.emitNewSigning(signing)
}

// handleEventSigningSuccess implements emitter handler for SigningSuccess event.
func (h *Hook) handleEventSigningSuccess(ctx sdk.Context, evMap common.EvMap) {
	id := tss.SigningID(common.Atoi(evMap[tsstypes.EventTypeSigningSuccess+"."+types.AttributeKeySigningID][0]))
	signing := h.tssKeeper.MustGetSigning(ctx, id)

	h.emitUpdateSigningSuccess(signing)
}

// handleEventSigningFailed implements emitter handler for SigningSuccess event.
func (h *Hook) handleEventSigningFailed(ctx sdk.Context, evMap common.EvMap) {
	id := tss.SigningID(common.Atoi(evMap[tsstypes.EventTypeSigningFailed+"."+types.AttributeKeySigningID][0]))
	signing := h.tssKeeper.MustGetSigning(ctx, id)

	if reason, ok := evMap[tsstypes.EventTypeSigningFailed+"."+tsstypes.AttributeKeyReason]; ok {
		h.emitUpdateSigningFailed(
			reason[0],
			signing,
		)
	} else {
		h.emitUpdateSigningFailed(
			"failed with on reason",
			signing,
		)
	}
}

// handleEventActivateTSSAccount implements emitter handler for tss account activate event.
func (h *Hook) handleEventActivateTSSAccount(ctx sdk.Context, evMap common.EvMap) {
	address := sdk.MustAccAddressFromBech32(evMap[tsstypes.EventTypeActivate+"."+tsstypes.AttributeKeyMember][0])
	status := h.tssKeeper.GetStatus(ctx, address)

	h.emitSetTSSAccountStatus(status)
}

// handleEventSetGroup implements emitter handler for events related to groups.
func (h *Hook) handleEventSetGroup(ctx sdk.Context, evMap common.EvMap) {
	gid := tss.GroupID(common.Atoi(evMap[tsstypes.EventTypeCreateGroup+"."+tsstypes.AttributeKeyGroupID][0]))
	group := h.tssKeeper.MustGetGroup(ctx, gid)
	dkgContext, err := h.tssKeeper.GetDKGContext(ctx, gid)
	if err != nil {
		panic(err)
	}

	h.emitSetGroup(group, dkgContext)

	members := h.tssKeeper.MustGetMembers(ctx, gid)
	for _, m := range members {
		h.emitSetMember(m)
	}
}
