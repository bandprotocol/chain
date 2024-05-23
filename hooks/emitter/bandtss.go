package emitter

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func (h *Hook) emitSetBandtssReplacement(replacement types.Replacement) {
	h.Write("SET_BAND_TSS_REPLACEMENT", common.JsDict{
		"tss_signing_id":   replacement.SigningID,
		"new_group_id":     replacement.NewGroupID,
		"new_pub_key":      parseBytes(replacement.NewPubKey),
		"current_group_id": replacement.CurrentGroupID,
		"current_pub_key":  parseBytes(replacement.CurrentPubKey),
		"exec_time":        replacement.ExecTime.UnixNano(),
		"status":           replacement.Status,
	})
}

func (h *Hook) emitSetBandtssGroup(gid tss.GroupID) {
	h.Write("SET_BAND_TSS_GROUP", common.JsDict{
		"current_group_id": gid,
		"since":            time.Now().UnixNano(),
	})
}

func (h *Hook) emitSetBandtssMember(member types.Member) {
	h.Write("SET_BAND_TSS_MEMBER", common.JsDict{
		"address":     member.Address,
		"is_active":   member.IsActive,
		"since":       member.Since.UnixNano(),
		"last_active": member.LastActive.UnixNano(),
	})
}

func (h *Hook) emitRemoveBandtssMember() {
	h.Write("REMOVE_BAND_TSS_MEMBERS", common.JsDict{})
}

func (h *Hook) emitSetBandtssSigning(signing types.Signing) {
	h.Write("SET_BAND_TSS_SIGNING", common.JsDict{
		"id":                         signing.ID,
		"fee":                        signing.Fee.String(),
		"requester":                  signing.Requester,
		"current_group_signing_id":   signing.CurrentGroupSigningID,
		"replacing_group_signing_id": signing.ReplacingGroupSigningID,
	})
}

// handleInitBandTSSModule implements emitter handler for init bandtss module.
func (h *Hook) handleInitBandTSSModule(ctx sdk.Context) {
	for _, signing := range h.bandtssKeeper.GetSignings(ctx) {
		h.emitSetBandtssSigning(signing)
	}
}

// handleNewBandtssGroupActive implements emitter handler for new bandtss group active.
func (h *Hook) handleNewBandtssGroupActive(ctx sdk.Context, gid tss.GroupID) {
	h.emitSetBandtssGroup(gid)
	h.emitRemoveBandtssMember()

	members := h.bandtssKeeper.GetMembers(ctx)
	for _, m := range members {
		h.emitSetBandtssMember(m)
	}
}

// handleUpdateBandtssStatus implements emitter handler for update bandtss status.
func (h *Hook) handleUpdateBandtssStatus(ctx sdk.Context, address sdk.AccAddress) {
	member, err := h.bandtssKeeper.GetMember(ctx, address)
	if err != nil {
		panic(err)
	}
	h.emitSetBandtssMember(member)
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

// handleBandtssMsgRequestSignature implements emitter handler for MsgRequestSignature of bandtss.
func (h *Hook) handleEventSigningRequestCreated(ctx sdk.Context, sid types.SigningID) {
	signing := h.bandtssKeeper.MustGetSigning(ctx, sid)
	h.emitSetBandtssSigning(signing)
}

// handleSetBandtssReplacement implements emitter handler events related to create replacements.
func (h *Hook) handleSetBandtssReplacement(ctx sdk.Context) {
	r := h.bandtssKeeper.GetReplacement(ctx)
	h.emitSetBandtssReplacement(r)
}
