package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func (h *Hook) emitNewBandtssGroupTransition(
	proposalID uint64,
	transition types.GroupTransition,
	createdHeight int64,
) {
	h.Write("NEW_BANDTSS_GROUP_TRANSITION", common.JsDict{
		"proposal_id":            proposalID,
		"tss_signing_id":         transition.SigningID,
		"current_tss_group_id":   transition.CurrentGroupID,
		"incoming_tss_group_id":  transition.IncomingGroupID,
		"current_group_pub_key":  parseBytes(transition.CurrentGroupPubKey),
		"incoming_group_pub_key": parseBytes(transition.IncomingGroupPubKey),
		"status":                 transition.Status,
		"exec_time":              transition.ExecTime.UnixNano(),
		"is_force_transition":    transition.IsForceTransition,
		"created_height":         createdHeight,
	})
}

func (h *Hook) emitUpdateBandtssGroupTransitionStatus(transition types.GroupTransition) {
	h.Write("UPDATE_BANDTSS_GROUP_TRANSITION", common.JsDict{
		"tss_signing_id":         transition.SigningID,
		"incoming_tss_group_id":  transition.IncomingGroupID,
		"incoming_group_pub_key": parseBytes(transition.IncomingGroupPubKey),
		"status":                 transition.Status,
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

func (h *Hook) emitNewBandtssCurrentGroup(gid tss.GroupID, transitionHeight int64) {
	h.Write("NEW_BANDTSS_CURRENT_GROUP", common.JsDict{
		"current_tss_group_id": gid,
		"transition_height":    transitionHeight,
	})
}

func (h *Hook) emitSetBandtssMember(member types.Member) {
	h.Write("SET_BANDTSS_MEMBER", common.JsDict{
		"address":      member.Address,
		"tss_group_id": member.GroupID,
		"is_active":    member.IsActive,
		"since":        member.Since.UnixNano(),
	})
}

func (h *Hook) emitNewBandtssSigning(signing types.Signing) {
	h.Write("New_BANDTSS_SIGNING", common.JsDict{
		"id":                            signing.ID,
		"fee_per_signer":                signing.FeePerSigner.String(),
		"requester":                     signing.Requester,
		"current_group_tss_signing_id":  signing.CurrentGroupSigningID,
		"incoming_group_tss_signing_id": signing.IncomingGroupSigningID,
	})
}

// handleInitBandTSSModule implements emitter handler for init bandtss module.
func (h *Hook) handleInitBandtssModule(ctx sdk.Context) {
	currentGroupID := h.bandtssKeeper.GetCurrentGroup(ctx).GroupID
	if currentGroupID != 0 {
		h.emitNewBandtssCurrentGroup(currentGroupID, ctx.BlockHeight())
	}

	members := h.bandtssKeeper.GetMembers(ctx)
	for _, m := range members {
		h.emitSetBandtssMember(m)
	}
}

// handleBandtssUpdateMember implements emitter handler for update bandtss status.
func (h *Hook) handleBandtssUpdateMember(ctx sdk.Context, address sdk.AccAddress, groupID tss.GroupID) {
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

	h.handleBandtssUpdateMember(ctx, acc, msg.GroupID)
}

// handleBandtssEventInactiveStatuses implements emitter handler for inactive status event.
func (h *Hook) handleBandtssEventInactiveStatuses(ctx sdk.Context, evMap common.EvMap) {
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
		h.handleBandtssUpdateMember(ctx, acc, groupID)
	}
}

// handleBandtssEventGroupTransition implements emitter handler for group transition event.
func (h *Hook) handleBandtssEventGroupTransition(ctx sdk.Context, eventIdx int, querier *EventQuerier) {
	// if transition not found, skip the process. There is a case that the transition message is signed
	// at the same block as the transition can be executed. The transition status will be updated via
	// another event (group_transition_success, group_transition_failed).
	transition, found := h.bandtssKeeper.GetGroupTransition(ctx)
	if !found {
		return
	}

	// set new bandtss members.
	if transition.Status == types.TRANSITION_STATUS_WAITING_EXECUTION {
		tssMembers := h.tssKeeper.MustGetMembers(ctx, transition.IncomingGroupID)
		for _, tssMember := range tssMembers {
			addr := sdk.MustAccAddressFromBech32(tssMember.Address)
			member, err := h.bandtssKeeper.GetMember(ctx, addr, transition.IncomingGroupID)
			if err != nil {
				panic(err)
			}
			h.emitSetBandtssMember(member)
		}
	}

	// check if it is a new transition or update the status of the existing transition.
	isNewTransition := transition.Status == types.TRANSITION_STATUS_CREATING_GROUP ||
		(transition.IsForceTransition && transition.Status == types.TRANSITION_STATUS_WAITING_EXECUTION)

	if isNewTransition {
		proposalID, found := getCurrentProposalID(eventIdx, querier)
		if !found {
			panic("proposal ID not found")
		}

		h.emitNewBandtssGroupTransition(proposalID, transition, ctx.BlockHeight())
	} else {
		h.emitUpdateBandtssGroupTransitionStatus(transition)
	}
}

// handleBandtssEventGroupTransitionSuccess implements emitter handler for group transition success event.
func (h *Hook) handleBandtssEventGroupTransitionSuccess(ctx sdk.Context, evMap common.EvMap) {
	// use value from emitted event due to the transition info is removed from the store.
	signingIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+tsstypes.AttributeKeySigningID]
	currentGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyCurrentGroupID]
	incomingGroupIDs := evMap[types.EventTypeGroupTransitionSuccess+"."+types.AttributeKeyIncomingGroupID]

	signingID := tss.SigningID(common.Atoi(signingIDs[0]))
	currentGroupID := tss.GroupID(common.Atoi(currentGroupIDs[0]))
	incomingGroupID := tss.GroupID(common.Atoi(incomingGroupIDs[0]))

	h.emitUpdateBandtssGroupTransitionStatusSuccess(types.GroupTransition{
		SigningID:       signingID,
		CurrentGroupID:  currentGroupID,
		IncomingGroupID: incomingGroupID,
	})

	h.emitNewBandtssCurrentGroup(incomingGroupID, ctx.BlockHeight())
}

// handleBandtssEventGroupTransitionFailed implements emitter handler for group transition failed event.
func (h *Hook) handleBandtssEventGroupTransitionFailed(_ sdk.Context, evMap common.EvMap) {
	// use value from emitted event due to the transition info is removed from the store.
	signingIDs := evMap[types.EventTypeGroupTransitionFailed+"."+tsstypes.AttributeKeySigningID]
	incomingGroupIDs := evMap[types.EventTypeGroupTransitionFailed+"."+types.AttributeKeyIncomingGroupID]
	currentGroupIDs := evMap[types.EventTypeGroupTransitionFailed+"."+types.AttributeKeyCurrentGroupID]

	h.emitUpdateBandtssGroupTransitionStatusFailed(types.GroupTransition{
		SigningID:       tss.SigningID(common.Atoi(signingIDs[0])),
		CurrentGroupID:  tss.GroupID(common.Atoi(currentGroupIDs[0])),
		IncomingGroupID: tss.GroupID(common.Atoi(incomingGroupIDs[0])),
	})
}

// handleBandtssEventSigningRequestCreated implements emitter handler for MsgRequestSignature of bandtss.
func (h *Hook) handleBandtssEventSigningRequestCreated(ctx sdk.Context, evMap common.EvMap) {
	signingIDs := evMap[types.EventTypeSigningRequestCreated+"."+types.AttributeKeySigningID]

	for _, sid := range signingIDs {
		signing := h.bandtssKeeper.MustGetSigning(ctx, types.SigningID(common.Atoui(sid)))
		h.emitNewBandtssSigning(signing)
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
			return common.Atoui(attr.Value), true
		}
	}

	return 0, false
}
