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

// handleInitTssModule implements emitter handler for initializing tss module.
func (h *Hook) handleInitTssModule(ctx sdk.Context) {
	for _, signing := range h.tssKeeper.GetSignings(ctx) {
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
