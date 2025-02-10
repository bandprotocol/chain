package emitter

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/bandprotocol/chain/v3/hooks/common"
)

var (
	EventTypeInactiveProposal = types.EventTypeInactiveProposal
	EventTypeActiveProposal   = types.EventTypeActiveProposal
	StatusInactive            = 6
)

func (h *Hook) emitSetDeposit(ctx sdk.Context, txHash []byte, id uint64, depositor sdk.AccAddress) {
	deposit, _ := h.govKeeper.Deposits.Get(ctx, collections.Join(id, depositor))
	h.Write("SET_DEPOSIT", common.JsDict{
		"proposal_id": id,
		"depositor":   depositor,
		"amount":      sdk.NewCoins(deposit.Amount...).String(),
		"tx_hash":     txHash,
	})
}

func (h *Hook) emitUpdateProposalAfterDeposit(ctx sdk.Context, id uint64) {
	proposal, _ := h.govKeeper.Proposals.Get(ctx, id)
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":              id,
		"status":          int(proposal.Status),
		"total_deposit":   sdk.NewCoins(proposal.TotalDeposit...).String(),
		"voting_time":     common.TimeToNano(proposal.VotingStartTime),
		"voting_end_time": common.TimeToNano(proposal.VotingEndTime),
	})
}

func (h *Hook) emitSetVoteWeighted(setVoteWeighted common.JsDict, options []*v1.WeightedVoteOption) {
	required_options := map[string]string{"yes": "0", "abstain": "0", "no": "0", "no_with_veto": "0"}

	for _, item := range options {
		switch item.Option {
		case v1.OptionYes:
			required_options["yes"] = item.Weight
		case v1.OptionAbstain:
			required_options["abstain"] = item.Weight
		case v1.OptionNo:
			required_options["no"] = item.Weight
		case v1.OptionNoWithVeto:
			required_options["no_with_veto"] = item.Weight
		}
	}

	for option, weight := range required_options {
		setVoteWeighted[option] = weight
	}
	h.Write("SET_VOTE_WEIGHTED", setVoteWeighted)
}

func (h *Hook) emitV1beta1SetVoteWeighted(setVoteWeighted common.JsDict, options []v1beta1.WeightedVoteOption) {
	required_options := map[string]string{"yes": "0", "abstain": "0", "no": "0", "no_with_veto": "0"}

	for _, item := range options {
		switch item.Option {
		case v1beta1.OptionYes:
			required_options["yes"] = item.Weight.String()
		case v1beta1.OptionAbstain:
			required_options["abstain"] = item.Weight.String()
		case v1beta1.OptionNo:
			required_options["no"] = item.Weight.String()
		case v1beta1.OptionNoWithVeto:
			required_options["no_with_veto"] = item.Weight.String()
		}
	}

	for option, weight := range required_options {
		setVoteWeighted[option] = weight
	}
	h.Write("SET_VOTE_WEIGHTED", setVoteWeighted)
}

// handleMsgSubmitProposal implements emitter handler for MsgSubmitProposal.
func (h *Hook) handleMsgSubmitProposal(
	ctx sdk.Context, txHash []byte, msg *v1.MsgSubmitProposal, evMap common.EvMap, detail common.JsDict,
) {
	proposalId := uint64(common.Atoi(evMap[types.EventTypeSubmitProposal+"."+types.AttributeKeyProposalID][0]))
	proposal, _ := h.govKeeper.Proposals.Get(ctx, proposalId)

	subMsg := proposal.Messages[0].GetCachedValue()
	switch subMsg := subMsg.(type) {
	case *v1.MsgExecLegacyContent:
		content := subMsg.Content.GetCachedValue().(v1beta1.Content)
		h.Write("NEW_PROPOSAL", common.JsDict{
			"id":               proposalId,
			"proposer":         msg.Proposer,
			"type":             content.ProposalType(),
			"title":            content.GetTitle(),
			"description":      content.GetDescription(),
			"proposal_route":   content.ProposalRoute(),
			"status":           int(proposal.Status),
			"submit_time":      common.TimeToNano(proposal.SubmitTime),
			"deposit_end_time": common.TimeToNano(proposal.DepositEndTime),
			"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
			"voting_time":      common.TimeToNano(proposal.VotingStartTime),
			"voting_end_time":  common.TimeToNano(proposal.VotingEndTime),
			"content":          content,
		})
	case sdk.Msg:
		h.Write("NEW_PROPOSAL", common.JsDict{
			"id":               proposalId,
			"proposer":         msg.Proposer,
			"type":             sdk.MsgTypeURL(subMsg),
			"title":            msg.Title,
			"description":      msg.Summary,
			"proposal_route":   sdk.MsgTypeURL(subMsg),
			"status":           int(proposal.Status),
			"submit_time":      common.TimeToNano(proposal.SubmitTime),
			"deposit_end_time": common.TimeToNano(proposal.DepositEndTime),
			"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
			"voting_time":      common.TimeToNano(proposal.VotingStartTime),
			"voting_end_time":  common.TimeToNano(proposal.VotingEndTime),
			"content":          subMsg,
		})
	default:
		break
	}

	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	h.emitSetDeposit(ctx, txHash, proposalId, proposer)
	detail["proposal_id"] = proposalId
}

// handleV1beta1MsgSubmitProposal implements emitter handler for MsgSubmitProposal v1beta1.
func (h *Hook) handleV1beta1MsgSubmitProposal(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgSubmitProposal, evMap common.EvMap, detail common.JsDict,
) {
	proposalId := uint64(common.Atoi(evMap[types.EventTypeSubmitProposal+"."+types.AttributeKeyProposalID][0]))
	proposal, _ := h.govKeeper.Proposals.Get(ctx, proposalId)
	content := msg.GetContent()

	h.Write("NEW_PROPOSAL", common.JsDict{
		"id":               proposalId,
		"proposer":         msg.Proposer,
		"type":             content.ProposalType(),
		"title":            content.GetTitle(),
		"description":      content.GetDescription(),
		"proposal_route":   content.ProposalRoute(),
		"status":           int(proposal.Status),
		"submit_time":      common.TimeToNano(proposal.SubmitTime),
		"deposit_end_time": common.TimeToNano(proposal.DepositEndTime),
		"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
		"voting_time":      common.TimeToNano(proposal.VotingStartTime),
		"voting_end_time":  common.TimeToNano(proposal.VotingEndTime),
		"content":          content,
	})
	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	h.emitSetDeposit(ctx, txHash, proposalId, proposer)
	detail["proposal_id"] = proposalId
}

// handleMsgDeposit implements emitter handler for MsgDeposit.
func (h *Hook) handleMsgDeposit(
	ctx sdk.Context, txHash []byte, msg *v1.MsgDeposit, detail common.JsDict,
) {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	h.emitSetDeposit(ctx, txHash, msg.ProposalId, depositor)
	h.emitUpdateProposalAfterDeposit(ctx, msg.ProposalId)
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)

	detail["title"] = proposal.Title
}

// handleV1beta1MsgDeposit implements emitter handler for MsgDeposit v1beta1.
func (h *Hook) handleV1beta1MsgDeposit(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgDeposit, detail common.JsDict,
) {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	h.emitSetDeposit(ctx, txHash, msg.ProposalId, depositor)
	h.emitUpdateProposalAfterDeposit(ctx, msg.ProposalId)
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)
	detail["title"] = proposal.Title
}

// handleMsgVote implements emitter handler for MsgVote.
func (h *Hook) handleMsgVote(
	ctx sdk.Context, txHash []byte, msg *v1.MsgVote, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitSetVoteWeighted(setVoteWeighted, v1.NewNonSplitVoteOption(msg.Option))
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)
	detail["title"] = proposal.Title
}

func (h *Hook) handleMsgCancelProposal(msg *v1.MsgCancelProposal) {
	h.Write("REMOVE_DEPOSIT", common.JsDict{
		"proposal_id": msg.ProposalId,
	})

	h.Write("REMOVE_VOTES", common.JsDict{
		"proposal_id": msg.ProposalId,
	})

	h.Write("REMOVE_PROPOSAL", common.JsDict{
		"id": msg.ProposalId,
	})
}

// handleV1beta1MsgVote implements emitter handler for MsgVote v1beta1.
func (h *Hook) handleV1beta1MsgVote(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgVote, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitV1beta1SetVoteWeighted(setVoteWeighted, v1beta1.NewNonSplitVoteOption(msg.Option))
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)
	detail["title"] = proposal.Title
}

// handleMsgVoteWeighted implements emitter handler for MsgVoteWeighted.
func (h *Hook) handleMsgVoteWeighted(
	ctx sdk.Context, txHash []byte, msg *v1.MsgVoteWeighted, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitSetVoteWeighted(setVoteWeighted, msg.Options)
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)
	detail["title"] = proposal.Title
}

// handleV1beta1MsgVoteWeighted implements emitter handler for MsgVoteWeighted v1beta1.
func (h *Hook) handleV1beta1MsgVoteWeighted(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgVoteWeighted, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitV1beta1SetVoteWeighted(setVoteWeighted, msg.Options)
	proposal, _ := h.govKeeper.Proposals.Get(ctx, msg.ProposalId)
	detail["title"] = proposal.Title
}

func (h *Hook) handleEventInactiveProposal(evMap common.EvMap) {
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":     common.Atoi(evMap[types.EventTypeInactiveProposal+"."+types.AttributeKeyProposalID][0]),
		"status": StatusInactive,
	})
}

func (h *Hook) handleEventTypeActiveProposal(ctx sdk.Context, evMap common.EvMap) {
	id := uint64(common.Atoi(evMap[types.EventTypeActiveProposal+"."+types.AttributeKeyProposalID][0]))
	proposal, _ := h.govKeeper.Proposals.Get(ctx, id)
	totalBond, _ := h.stakingKeeper.TotalBondedTokens(ctx)
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":                  id,
		"status":              int(proposal.Status),
		"total_bonded_tokens": totalBond,
		"yes_vote":            proposal.FinalTallyResult.YesCount,
		"no_vote":             proposal.FinalTallyResult.NoCount,
		"no_with_veto_vote":   proposal.FinalTallyResult.NoWithVetoCount,
		"abstain_vote":        proposal.FinalTallyResult.AbstainCount,
	})
}
