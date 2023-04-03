package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/bandprotocol/chain/v2/hooks/common"
)

var (
	EventTypeInactiveProposal = types.EventTypeInactiveProposal
	EventTypeActiveProposal   = types.EventTypeActiveProposal
	StatusInactive            = 6
)

func (h *Hook) emitSetDeposit(ctx sdk.Context, txHash []byte, id uint64, depositor sdk.AccAddress) {
	deposit, _ := h.govKeeper.GetDeposit(ctx, id, depositor)
	h.Write("SET_DEPOSIT", common.JsDict{
		"proposal_id": id,
		"depositor":   depositor,
		"amount":      sdk.NewCoins(deposit.Amount...).String(),
		"tx_hash":     txHash,
	})
}

func (h *Hook) emitUpdateProposalAfterDeposit(ctx sdk.Context, id uint64) {
	proposal, _ := h.govKeeper.GetProposal(ctx, id)
	h.Write("UPDATE_PROPOSAL", common.JsDict{
		"id":              id,
		"status":          int(proposal.Status),
		"total_deposit":   sdk.NewCoins(proposal.TotalDeposit...).String(),
		"voting_time":     proposal.VotingStartTime.UnixNano(),
		"voting_end_time": proposal.VotingEndTime.UnixNano(),
	})
}

func (h *Hook) emitSetVoteWeighted(setVoteWeighted common.JsDict, options []v1beta1.WeightedVoteOption) {
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
func (app *Hook) handleMsgSubmitProposal(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgSubmitProposal, evMap common.EvMap, detail common.JsDict,
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
		"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
		"voting_time":      proposal.VotingStartTime.UnixNano(),
		"voting_end_time":  proposal.VotingEndTime.UnixNano(),
		"content":          content,
	})
	proposer, _ := sdk.AccAddressFromBech32(msg.Proposer)
	app.emitSetDeposit(ctx, txHash, proposalId, proposer)
	detail["proposal_id"] = proposalId
}

// handleMsgDeposit implements emitter handler for MsgDeposit.
func (h *Hook) handleMsgDeposit(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgDeposit, detail common.JsDict,
) {
	depositor, _ := sdk.AccAddressFromBech32(msg.Depositor)
	h.emitSetDeposit(ctx, txHash, msg.ProposalId, depositor)
	h.emitUpdateProposalAfterDeposit(ctx, msg.ProposalId)
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetTitle()
}

// handleMsgVote implements emitter handler for MsgVote.
func (h *Hook) handleMsgVote(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgVote, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitSetVoteWeighted(setVoteWeighted, v1beta1.NewNonSplitVoteOption(msg.Option))
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetTitle()
}

// handleMsgVoteWeighted implements emitter handler for MsgVoteWeighted.
func (h *Hook) handleMsgVoteWeighted(
	ctx sdk.Context, txHash []byte, msg *v1beta1.MsgVoteWeighted, detail common.JsDict,
) {
	setVoteWeighted := common.JsDict{
		"proposal_id": msg.ProposalId,
		"voter":       msg.Voter,
		"tx_hash":     txHash,
	}
	h.emitSetVoteWeighted(setVoteWeighted, msg.Options)
	proposal, _ := h.govKeeper.GetProposal(ctx, msg.ProposalId)
	detail["title"] = proposal.Messages[0].GetCachedValue().(*v1.MsgExecLegacyContent).Content.GetCachedValue().(v1beta1.Content).GetTitle()
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
		"id":                  id,
		"status":              int(proposal.Status),
		"total_bonded_tokens": h.stakingKeeper.TotalBondedTokens(ctx),
		"yes_vote":            proposal.FinalTallyResult.YesCount,
		"no_vote":             proposal.FinalTallyResult.NoCount,
		"no_with_veto_vote":   proposal.FinalTallyResult.NoWithVetoCount,
		"abstain_vote":        proposal.FinalTallyResult.AbstainCount,
	})
}
