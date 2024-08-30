package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func (h *Hook) emitSetBandtssGroupTransition(
	proposalID uint64,
	transition types.GroupTransition,
	createdHeight int64,
) {
	h.Write("SET_BANDTSS_GROUP_TRANSITION", common.JsDict{
		"proposal_id":            proposalID,
		"tss_signing_id":         transition.SigningID,
		"current_tss_group_id":   transition.CurrentGroupID,
		"incoming_tss_group_id":  transition.IncomingGroupID,
		"current_group_pub_key":  parseBytes(transition.CurrentGroupPubKey),
		"incoming_group_pub_key": parseBytes(transition.IncomingGroupPubKey),
		"status":                 transition.Status,
		"exec_time":              transition.ExecTime.UnixNano(),
		"created_height":         createdHeight,
	})
}

func (h *Hook) emitUpdateBandtssGroupTransitionStatus(transition types.GroupTransition) {
	h.Write("UPDATE_BANDTSS_GROUP_TRANSITION", common.JsDict{
		"tss_signing_id":        transition.SigningID,
		"current_tss_group_id":  transition.CurrentGroupID,
		"incoming_tss_group_id": transition.IncomingGroupID,
		"status":                transition.Status,
	})
}

func (h *Hook) emitUpdateBandtssGroupTransitionStatusSuccess(transition types.GroupTransition) {
	h.Write("UPDATE_BANDTSS_GROUP_TRANSITION_SUCCESS", common.JsDict{
		"tss_signing_id":        transition.SigningID,
		"current_tss_group_id":  transition.CurrentGroupID,
		"incoming_tss_group_id": transition.IncomingGroupID,
	})
}

func (h *Hook) emitUpdateBandtssGroupTransitionStatusFailed(transition types.GroupTransition) {
	h.Write("UPDATE_BANDTSS_GROUP_TRANSITION_FAILED", common.JsDict{
		"tss_signing_id":        transition.SigningID,
		"current_tss_group_id":  transition.CurrentGroupID,
		"incoming_tss_group_id": transition.IncomingGroupID,
	})
}

func (h *Hook) emitSetBandtssCurrentGroup(gid tss.GroupID, transitionHeight int64) {
	h.Write("SET_BANDTSS_CURRENT_GROUP", common.JsDict{
		"current_tss_group_id": gid,
		"transition_height":    transitionHeight,
	})
}

func (h *Hook) emitSetBandtssMember(member types.Member) {
	h.Write("SET_BANDTSS_MEMBER", common.JsDict{
		"address":       member.Address,
		"tss_group_id":  member.GroupID,
		"is_active":     member.IsActive,
		"penalty_since": member.Since.UnixNano(),
		"last_active":   member.LastActive.UnixNano(),
	})
}

func (h *Hook) emitSetBandtssSigning(signing types.Signing) {
	h.Write("SET_BANDTSS_SIGNING", common.JsDict{
		"id":                             signing.ID,
		"fee_per_signer":                 signing.FeePerSigner.String(),
		"requester":                      signing.Requester,
		"current_group_tss_signing_id":   signing.CurrentGroupSigningID,
		"replacing_group_tss_signing_id": signing.IncomingGroupSigningID,
	})
}

// handleInitBandTSSModule implements emitter handler for init bandtss module.
func (h *Hook) handleInitBandtssModule(ctx sdk.Context) {
	currentGroupID := h.bandtssKeeper.GetCurrentGroupID(ctx)
	if currentGroupID != 0 {
		h.emitSetBandtssCurrentGroup(currentGroupID, ctx.BlockHeight())
	}

	members := h.bandtssKeeper.GetMembers(ctx)
	for _, m := range members {
		h.emitSetBandtssMember(m)
	}
}

// handleUpdateBandtssMember implements emitter handler for update bandtss status.
func (h *Hook) handleUpdateBandtssMember(ctx sdk.Context, address sdk.AccAddress, groupID tss.GroupID) {
	member, err := h.bandtssKeeper.GetMember(ctx, address, groupID)
	if err != nil {
		panic(err)
	}
	h.emitSetBandtssMember(member)
}

// handleBandtssMsgActivate implements emitter handler for MsgActivate of bandtss.
func (h *Hook) handleBandtssMsgActivate(ctx sdk.Context, msg *types.MsgActivate) {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	h.handleUpdateBandtssMember(ctx, acc, msg.GroupID)
}

// handleBandtssMsgHeartbeat implements emitter handler for MsgHeartbeat of bandtss.
func (h *Hook) handleBandtssMsgHeartbeat(ctx sdk.Context, msg *types.MsgHeartbeat) {
	acc, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}

	h.handleUpdateBandtssMember(ctx, acc, msg.GroupID)
}

// handleEventInactiveStatuses implements emitter handler for inactive status event.
func (h *Hook) handleEventInactiveStatuses(ctx sdk.Context, evMap common.EvMap) {
	addresses := evMap[types.EventTypeInactiveStatus+"."+types.AttributeKeyAddress]
	groupIDs := evMap[types.EventTypeInactiveStatus+"."+types.AttributeKeyGroupID]
	if len(addresses) != len(groupIDs) {
		panic("invalid event data")
	}

	for i, addr := range addresses {
		acc, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			panic(err)
		}

		groupID := tss.GroupID(common.Atoi(groupIDs[i]))
		h.handleUpdateBandtssMember(ctx, acc, groupID)
	}
}

// handleEventGroupTransition implements emitter handler for group transition event.
func (h *Hook) handleEventGroupTransition(ctx sdk.Context, eventIdx int, querier *EventQuerier) {
	// if transition not found, skip the process. There is a case that the transition message is signed
	// at the same block as the transition can be executed. The transition status will be updated via
	// another event (group_transition_success, group_transition_failed).
	transition, found := h.bandtssKeeper.GetGroupTransition(ctx)
	if !found {
		return
	}

	isNewTransition := transition.Status == types.TRANSITION_STATUS_CREATING_GROUP ||
		(transition.Status == types.TRANSITION_STATUS_WAITING_SIGN && transition.SigningID == 0)

	if isNewTransition {
		proposalID, _ := getCurrentProposalID(eventIdx, querier)
		h.emitSetBandtssGroupTransition(proposalID, transition, ctx.BlockHeight())
	} else {
		h.emitUpdateBandtssGroupTransitionStatus(transition)
	}
}

// handleEventGroupTransitionSuccess implements emitter handler for group transition success event.
func (h *Hook) handleEventGroupTransitionSuccess(ctx sdk.Context, evMap common.EvMap) {
	// use value from emitted event due to the transition info is removed from the store.
	signingIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeySigningID]
	incomingGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyIncomingGroupID]
	currentGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyCurrentGroupID]

	signingID := tss.SigningID(common.Atoi(signingIDs[0]))
	currentGroupID := tss.GroupID(common.Atoi(currentGroupIDs[0]))
	incomingGroupID := tss.GroupID(common.Atoi(incomingGroupIDs[0]))

	h.emitUpdateBandtssGroupTransitionStatusSuccess(types.GroupTransition{
		SigningID:       signingID,
		CurrentGroupID:  currentGroupID,
		IncomingGroupID: incomingGroupID,
	})

	h.emitSetBandtssCurrentGroup(currentGroupID, ctx.BlockHeight())
}

// handleEventGroupTransitionFailed implements emitter handler for group transition failed event.
func (h *Hook) handleEventGroupTransitionFailed(_ sdk.Context, evMap common.EvMap) {
	// use value from emitted event due to the transition info is removed from the store.
	signingIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeySigningID]
	incomingGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyIncomingGroupID]
	currentGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyCurrentGroupID]

	h.emitUpdateBandtssGroupTransitionStatusFailed(types.GroupTransition{
		SigningID:       tss.SigningID(common.Atoi(signingIDs[0])),
		CurrentGroupID:  tss.GroupID(common.Atoi(currentGroupIDs[0])),
		IncomingGroupID: tss.GroupID(common.Atoi(incomingGroupIDs[0])),
	})
}

// handleBandtssMsgRequestSignature implements emitter handler for MsgRequestSignature of bandtss.
func (h *Hook) handleEventSigningRequestCreated(ctx sdk.Context, evMap common.EvMap) {
	signingIDs := evMap[types.EventTypeSigningRequestCreated+"."+types.AttributeKeySigningID]

	for _, sid := range signingIDs {
		signing := h.bandtssKeeper.MustGetSigning(ctx, types.SigningID(uint64(common.Atoi(sid))))
		h.emitSetBandtssSigning(signing)
	}
}

// getCurrentProposalID returns the proposal ID that execute the process that emit the given event.
// If the event is triggered by the proposal, the active proposal event should be emitted next to the event.
func getCurrentProposalID(eventIdx int, querier *EventQuerier) (id uint64, found bool) {
	proposalIDEvent, found := querier.FindEventWithTypeAfterIdx(govtypes.EventTypeActiveProposal, eventIdx)
	if !found {
		return 0, false
	}

	for _, attr := range proposalIDEvent.Attributes {
		if attr.Key == govtypes.AttributeKeyProposalID {
			return uint64(common.Atoi(attr.Value)), true
		}
	}

	return 0, false
}
