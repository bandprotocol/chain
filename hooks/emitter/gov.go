package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/hooks/common"
)

var (
	EventTypeInactiveProposal = types.EventTypeInactiveProposal
	EventTypeActiveProposal   = types.EventTypeActiveProposal
	StatusInactive            = 6
)

func (h *Hook) emitGovModule(ctx sdk.Context) {
	h.govKeeper.IterateProposals(ctx, func(proposal types.Proposal) (stop bool) {
		h.emitNewProposal(proposal, nil)
		return false
	})
	h.govKeeper.IterateAllDeposits(ctx, func(deposit types.Deposit) (stop bool) {
		h.Write("SET_DEPOSIT", common.JsDict{
			"proposal_id": deposit.ProposalId,
			"depositor":   deposit.Depositor,
			"amount":      deposit.Amount.String(),
			"tx_hash":     nil,
		})
		return false
	})
	h.govKeeper.IterateAllVotes(ctx, func(vote types.Vote) (stop bool) {
		h.Write("SET_VOTE", common.JsDict{
			"proposal_id": vote.ProposalId,
			"voter":       vote.Voter,
			"answer":      int(vote.Option),
			"tx_hash":     nil,
		})
		return false
	})
}

func (h *Hook) emitNewProposal(proposal types.Proposal, proposer sdk.AccAddress) {
	content := proposal.GetContent()
	h.Write("NEW_PROPOSAL", common.JsDict{
		"id":               proposal.ProposalId,
		"proposer":         proposer,
		"type":             content.ProposalType(),
		"title":            content.GetTitle(),
		"description":      content.GetDescription(),
		"proposal_route":   content.ProposalRoute(),
		"status":           int(proposal.Status),
		"submit_time":      proposal.SubmitTime.UnixNano(),
		"deposit_end_time": proposal.DepositEndTime.UnixNano(),
		"total_deposit":    proposal.TotalDeposit.String(),
		"voting_time":      proposal.VotingStartTime.UnixNano(),
		"voting_end_time":  proposal.VotingEndTime.UnixNano(),
	})
}

func (h *Hook) emitSetDeposit(ctx sdk.Context, txHash []byte, id uint64, depositor sdk.AccAddress) {
	deposit, _ := h.govKeeper.GetDeposit(ctx, id, depositor)
	h.Write("SET_DEPOSIT", common.JsDict{
		"proposal_id": id,
		"depositor":   depositor,
		"amount":      deposit.Amount.String(),
		"tx_hash":     txHash,
	})
}

func (h *Hook) emitUpdateProposalAfterDeposit(ctx sdk.Context, id uint64) {
	proposal, _ := h.govKeeper.GetProposal(ctx, id)
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":              id,
		"status":          int(proposal.Status),
		"total_deposit":   proposal.TotalDeposit.String(),
		"voting_time":     proposal.VotingStartTime.UnixNano(),
		"voting_end_time": proposal.VotingEndTime.UnixNano(),
	})
}

// handleMsgSubmitProposal implements emitter handler for MsgSubmitProposal.
func (app *Hook) handleMsgSubmitProposal(
	ctx sdk.Context, txHash []byte, msg *types.MsgSubmitProposal, evMap common.EvMap, detail common.JsDict,
) {
	proposalId := uint64(common.Atoi(evMap[types.EventTypeSubmitProposal+"."+types.AttributeKeyProposalID][0]))
	proposal, _ := app.govKeeper.GetProposal(ctx, proposalId)
	content := msg.GetContent()
	app.Write("NEW_PROPOSAL", common.JsDict{
		"id":               proposalId,
		"proposer":         msg.Proposer,
		"type":             content.ProposalType(),
		"title":            content.GetTitle(),
		"description":      content.GetDescription(),
		"proposal_route":   content.ProposalRoute(),
		"status":           int(proposal.Status),
		"submit_time":      proposal.SubmitTime.UnixNano(),
		"deposit_end_time": proposal.DepositEndTime.UnixNano(),
		"total_deposit":    proposal.TotalDeposit.String(),
		"voting_time":      proposal.VotingStartTime.UnixNano(),
		"voting_end_time":  proposal.VotingEndTime.UnixNano(),
	})
	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	app.emitSetDeposit(ctx, txHash, proposalId, proposer)
	detail["proposal_id"] = proposalId
}

// handleMsgDeposit implements emitter handler for MsgDeposit.
func (h *Hook) handleMsgDeposit(
	ctx sdk.Context, txHash []byte, msg *types.MsgDeposit, detail common.JsDict,
) {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	h.emitSetDeposit(ctx, txHash, msg.ProposalId, depositor)
	h.emitUpdateProposalAfterDeposit(ctx, msg.ProposalId)
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.GetTitle()
}

// handleMsgVote implements emitter handler for MsgVote.
func (h *Hook) handleMsgVote(
	ctx sdk.Context, txHash []byte, msg *types.MsgVote, detail common.JsDict,
) {
	h.Write("SET_VOTE_WEIGHTED", common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"options":     types.NewNonSplitVoteOption(msg.Option),
		"tx_hash":     txHash,
	})
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.GetTitle()

}

// handleMsgVote implements emitter handler for MsgVote.
func (h *Hook) handleMsgVoteWeighted(
	ctx sdk.Context, txHash []byte, msg *types.MsgVoteWeighted, detail common.JsDict,
) {
	h.Write("SET_VOTE_WEIGHTED", common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"options":     msg.Options,
		"tx_hash":     txHash,
	})
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.GetTitle()

}

func (h *Hook) handleEventInactiveProposal(evMap common.EvMap) {
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":     common.Atoi(evMap[types.EventTypeInactiveProposal+"."+types.AttributeKeyProposalID][0]),
		"status": StatusInactive,
	})
}

func (h *Hook) handleEventTypeActiveProposal(ctx sdk.Context, evMap common.EvMap) {
	id := uint64(common.Atoi(evMap[types.EventTypeActiveProposal+"."+types.AttributeKeyProposalID][0]))
	proposal, _ := h.govKeeper.GetProposal(ctx, id)
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":     id,
		"status": int(proposal.Status),
	})
}
