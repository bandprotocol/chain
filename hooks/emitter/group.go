package emitter

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/group"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/bandprotocol/chain/v3/hooks/common"
)

func extractStringFromEventMap(evMap common.EvMap, event string, topic string) string {
	return strings.Trim(evMap[event+"."+topic][0], `"`)
}

// handleGroupMsgCreateGroup implements emitter handler for Group's MsgCreateGroup.
func (h *Hook) handleGroupMsgCreateGroup(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventCreateGroup{}), "group_id")),
	)
	groupInfoResponse, _ := h.groupKeeper.GroupInfo(
		sdk.WrapSDKContext(ctx),
		&group.QueryGroupInfoRequest{GroupId: groupId},
	)
	groupInfo := groupInfoResponse.Info
	h.Write("NEW_GROUP", common.JsDict{
		"id":           groupId,
		"version":      groupInfo.Version,
		"admin":        groupInfo.Admin,
		"metadata":     groupInfo.Metadata,
		"total_weight": groupInfo.TotalWeight,
		"created_at":   common.TimeToNano(&groupInfo.CreatedAt),
	})
	h.doAddGroupMembers(ctx, groupId)
}

// handleGroupMsgCreateGroup implements emitter handler for Group's MsgCreateGroupPolicy.
func (h *Hook) handleGroupMsgCreateGroupPolicy(
	ctx sdk.Context, evMap common.EvMap,
) {
	policyAddress := extractStringFromEventMap(evMap, proto.MessageName(&group.EventCreateGroupPolicy{}), "address")
	groupPolicyResponse, _ := h.groupKeeper.GroupPolicyInfo(
		sdk.WrapSDKContext(ctx),
		&group.QueryGroupPolicyInfoRequest{
			Address: policyAddress,
		},
	)
	groupPolicyInfo := groupPolicyResponse.Info
	decisionPolicy, _ := groupPolicyInfo.GetDecisionPolicy()
	h.Write("NEW_GROUP_POLICY", common.JsDict{
		"address":         policyAddress,
		"type":            proto.MessageName(decisionPolicy),
		"group_id":        groupPolicyInfo.GroupId,
		"admin":           groupPolicyInfo.Admin,
		"metadata":        groupPolicyInfo.Metadata,
		"version":         groupPolicyInfo.Version,
		"decision_policy": decisionPolicy,
		"created_at":      common.TimeToNano(&groupPolicyInfo.CreatedAt),
	})
}

// handleGroupMsgCreateGroupWithPolicy implements emitter handler for Group's MsgCreateGroupWithPolicy.
func (h *Hook) handleGroupMsgCreateGroupWithPolicy(
	ctx sdk.Context, evMap common.EvMap,
) {
	h.handleGroupMsgCreateGroup(ctx, evMap)
	h.handleGroupMsgCreateGroupPolicy(ctx, evMap)
}

// handleGroupMsgSubmitProposal implements emitter handler for Group's MsgSubmitProposal.
func (h *Hook) handleGroupMsgSubmitProposal(
	ctx sdk.Context, evMap common.EvMap,
) {
	proposalId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventSubmitProposal{}), "proposal_id")),
	)
	proposalResponse, _ := h.groupKeeper.Proposal(
		sdk.WrapSDKContext(ctx),
		&group.QueryProposalRequest{ProposalId: proposalId},
	)
	proposal := proposalResponse.Proposal
	msgs, _ := proposal.GetMsgs()
	messages := make([]common.JsDict, len(msgs))
	for i, m := range msgs {
		messages[i] = common.JsDict{
			"msg":  m,
			"type": sdk.MsgTypeURL(m),
		}
	}

	h.Write("NEW_GROUP_PROPOSAL", common.JsDict{
		"id":                   proposal.Id,
		"group_policy_address": proposal.GroupPolicyAddress,
		"metadata":             proposal.Metadata,
		"proposers":            strings.Join(proposal.Proposers, ","),
		"submit_time":          common.TimeToNano(&proposal.SubmitTime),
		"group_version":        proposal.GroupVersion,
		"group_policy_version": proposal.GroupPolicyVersion,
		"status":               proposal.Status.String(),
		"yes_vote":             proposal.FinalTallyResult.YesCount,
		"no_vote":              proposal.FinalTallyResult.NoCount,
		"no_with_veto_vote":    proposal.FinalTallyResult.NoWithVetoCount,
		"abstain_vote":         proposal.FinalTallyResult.AbstainCount,
		"voting_period_end":    common.TimeToNano(&proposal.VotingPeriodEnd),
		"executor_result":      proposal.ExecutorResult.String(),
		"messages":             messages,
		"title":                proposal.Title,
		"summary":              proposal.Summary,
	})
}

// handleGroupMsgVote implements emitter handler for Group's MsgVote.
func (h *Hook) handleGroupMsgVote(
	ctx sdk.Context, msg *group.MsgVote, evMap common.EvMap,
) {
	proposalId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventVote{}), "proposal_id")),
	)
	voteResponse, err := h.groupKeeper.VoteByProposalVoter(
		sdk.WrapSDKContext(ctx),
		&group.QueryVoteByProposalVoterRequest{
			ProposalId: proposalId,
			Voter:      msg.Voter,
		},
	)
	if err != nil {
		return
	}
	vote := voteResponse.Vote
	h.Write("NEW_GROUP_VOTE", common.JsDict{
		"group_proposal_id": proposalId,
		"voter_address":     vote.Voter,
		"option":            vote.Option.String(),
		"metadata":          vote.Metadata,
		"submit_time":       common.TimeToNano(&vote.SubmitTime),
	})
}

// handleGroupMsgLeaveGroup implements emitter handler for Group's MsgLeaveGroup.
func (h *Hook) handleGroupMsgLeaveGroup(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventLeaveGroup{}), "group_id")),
	)
	address := extractStringFromEventMap(evMap, proto.MessageName(&group.EventLeaveGroup{}), "address")
	h.doUpdateGroup(ctx, groupId)
	h.Write("REMOVE_GROUP_MEMBER", common.JsDict{
		"group_id": groupId,
		"address":  address,
	})
}

// handleGroupMsgUpdateGroupAdmin implements emitter handler for Group's MsgUpdateGroupAdmin.
func (h *Hook) handleGroupMsgUpdateGroupAdmin(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventUpdateGroup{}), "group_id")),
	)
	h.doUpdateGroup(ctx, groupId)
}

// handleGroupMsgUpdateGroupMembers implements emitter handler for Group's MsgUpdateGroupMembers.
func (h *Hook) handleGroupMsgUpdateGroupMembers(
	ctx sdk.Context, msg *group.MsgUpdateGroupMembers, evMap common.EvMap,
) {
	h.Write("REMOVE_GROUP_MEMBERS_BY_GROUP_ID", common.JsDict{
		"group_id": msg.GroupId,
	})
	h.doAddGroupMembers(ctx, msg.GroupId)
	h.doUpdateGroup(ctx, msg.GroupId)
}

// handleGroupMsgUpdateGroupMetadata implements emitter handler for Group's MsgUpdateGroupMetadata.
func (h *Hook) handleGroupMsgUpdateGroupMetadata(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventUpdateGroup{}), "group_id")),
	)
	h.doUpdateGroup(ctx, groupId)
}

// handleGroupMsgUpdateGroupPolicyAdmin implements emitter handler for Group's MsgUpdateGroupPolicyAdmin.
func (h *Hook) handleGroupMsgUpdateGroupPolicyAdmin(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupPolicyAddress := extractStringFromEventMap(
		evMap,
		proto.MessageName(&group.EventUpdateGroupPolicy{}),
		"address",
	)
	h.doUpdateGroupPolicy(ctx, groupPolicyAddress)
}

// handleGroupMsgUpdateGroupPolicyDecisionPolicy implements emitter handler for Group's MsgUpdateGroupPolicyDecisionPolicy.
func (h *Hook) handleGroupMsgUpdateGroupPolicyDecisionPolicy(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupPolicyAddress := extractStringFromEventMap(
		evMap,
		proto.MessageName(&group.EventUpdateGroupPolicy{}),
		"address",
	)
	h.doUpdateGroupPolicy(ctx, groupPolicyAddress)
}

// handleGroupMsgUpdateGroupPolicyMetadata implements emitter handler for Group's MsgUpdateGroupPolicyMetadata.
func (h *Hook) handleGroupMsgUpdateGroupPolicyMetadata(
	ctx sdk.Context, evMap common.EvMap,
) {
	groupPolicyAddress := extractStringFromEventMap(
		evMap,
		proto.MessageName(&group.EventUpdateGroupPolicy{}),
		"address",
	)
	h.doUpdateGroupPolicy(ctx, groupPolicyAddress)
}

// handleGroupMsgWithdrawProposal implements emitter handler for Group's MsgWithdrawProposal.
func (h *Hook) handleGroupMsgWithdrawProposal(
	ctx sdk.Context, evMap common.EvMap,
) {
	proposalId := uint64(
		common.Atoi(
			extractStringFromEventMap(evMap, proto.MessageName(&group.EventWithdrawProposal{}), "proposal_id"),
		),
	)
	h.doUpdateGroupProposal(ctx, proposalId)
}

// handleGroupEventExec implements emitter handler for Group's EventExec.
func (h *Hook) handleGroupEventExec(
	ctx sdk.Context, evMap common.EvMap,
) {
	if len(evMap[proto.MessageName(&group.EventExec{})+".proposal_id"]) == 0 {
		return
	}
	proposalId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventExec{}), "proposal_id")),
	)
	executorResult := extractStringFromEventMap(evMap, proto.MessageName(&group.EventExec{}), "result")
	h.Write("UPDATE_GROUP_PROPOSAL_BY_ID", common.JsDict{
		"id":              proposalId,
		"executor_result": executorResult,
	})

	h.handleGroupEventProposalPruned(ctx, evMap)
}

// handleGroupEventProposalPruned implements emitter handler for Group's EventProposalPruned.
func (h *Hook) handleGroupEventProposalPruned(
	ctx sdk.Context, evMap common.EvMap,
) {
	if len(evMap[proto.MessageName(&group.EventProposalPruned{})+".proposal_id"]) == 0 {
		return
	}
	proposalId := uint64(
		common.Atoi(extractStringFromEventMap(evMap, proto.MessageName(&group.EventProposalPruned{}), "proposal_id")),
	)
	proposalStatus := extractStringFromEventMap(evMap, proto.MessageName(&group.EventProposalPruned{}), "status")
	tallyResult := group.DefaultTallyResult()
	_ = json.Unmarshal([]byte(evMap[proto.MessageName(&group.EventProposalPruned{})+".tally_result"][0]), &tallyResult)
	h.Write("UPDATE_GROUP_PROPOSAL_BY_ID", common.JsDict{
		"id":                proposalId,
		"status":            proposalStatus,
		"yes_vote":          tallyResult.YesCount,
		"no_vote":           tallyResult.NoCount,
		"no_with_veto_vote": tallyResult.NoWithVetoCount,
		"abstain_vote":      tallyResult.AbstainCount,
	})
}

func (h *Hook) doUpdateGroup(ctx sdk.Context, groupId uint64) {
	groupInfoResponse, _ := h.groupKeeper.GroupInfo(
		sdk.WrapSDKContext(ctx),
		&group.QueryGroupInfoRequest{GroupId: groupId},
	)
	groupInfo := groupInfoResponse.Info
	h.Write("UPDATE_GROUP", common.JsDict{
		"id":           groupId,
		"version":      groupInfo.Version,
		"admin":        groupInfo.Admin,
		"metadata":     groupInfo.Metadata,
		"total_weight": groupInfo.TotalWeight,
		"created_at":   common.TimeToNano(&groupInfo.CreatedAt),
	})
}

func (h *Hook) doUpdateGroupPolicy(ctx sdk.Context, policyAddress string) {
	groupPolicyResponse, _ := h.groupKeeper.GroupPolicyInfo(
		sdk.WrapSDKContext(ctx),
		&group.QueryGroupPolicyInfoRequest{
			Address: policyAddress,
		},
	)
	groupPolicyInfo := groupPolicyResponse.Info
	decisionPolicy, _ := groupPolicyInfo.GetDecisionPolicy()
	h.Write("UPDATE_GROUP_POLICY", common.JsDict{
		"address":         policyAddress,
		"group_id":        groupPolicyInfo.GroupId,
		"admin":           groupPolicyInfo.Admin,
		"metadata":        groupPolicyInfo.Metadata,
		"version":         groupPolicyInfo.Version,
		"decision_policy": decisionPolicy,
		"created_at":      common.TimeToNano(&groupPolicyInfo.CreatedAt),
	})

	h.doAbortProposals(ctx, policyAddress)
}

func (h *Hook) doAbortProposals(ctx sdk.Context, policyAddress string) {
	groupProposalsResponse, _ := h.groupKeeper.ProposalsByGroupPolicy(
		sdk.WrapSDKContext(ctx),
		&group.QueryProposalsByGroupPolicyRequest{
			Address: policyAddress,
		},
	)
	for {
		groupProposals := groupProposalsResponse.Proposals
		for _, groupProposal := range groupProposals {
			if groupProposal.Status == group.PROPOSAL_STATUS_ABORTED {
				h.doUpdateGroupProposal(ctx, groupProposal.Id)
			}
		}
		if len(groupProposalsResponse.Pagination.NextKey) == 0 {
			break
		}
		groupProposalsResponse, _ = h.groupKeeper.ProposalsByGroupPolicy(
			sdk.WrapSDKContext(ctx),
			&group.QueryProposalsByGroupPolicyRequest{
				Address: policyAddress,
				Pagination: &query.PageRequest{
					Key: groupProposalsResponse.Pagination.NextKey,
				},
			},
		)
	}
}

func (h *Hook) doUpdateGroupProposal(ctx sdk.Context, proposalId uint64) {
	proposalResponse, _ := h.groupKeeper.Proposal(
		sdk.WrapSDKContext(ctx),
		&group.QueryProposalRequest{ProposalId: proposalId},
	)
	proposal := proposalResponse.Proposal
	msgs, _ := proposal.GetMsgs()
	messages := make([]common.JsDict, len(msgs))
	for i, m := range msgs {
		messages[i] = common.JsDict{
			"msg":  m,
			"type": sdk.MsgTypeURL(m),
		}
	}

	h.Write("UPDATE_GROUP_PROPOSAL", common.JsDict{
		"id":                   proposal.Id,
		"group_policy_address": proposal.GroupPolicyAddress,
		"metadata":             proposal.Metadata,
		"proposers":            strings.Join(proposal.Proposers, ","),
		"submit_time":          common.TimeToNano(&proposal.SubmitTime),
		"group_version":        proposal.GroupVersion,
		"group_policy_version": proposal.GroupPolicyVersion,
		"status":               proposal.Status.String(),
		"yes_vote":             proposal.FinalTallyResult.YesCount,
		"no_vote":              proposal.FinalTallyResult.NoCount,
		"no_with_veto_vote":    proposal.FinalTallyResult.NoWithVetoCount,
		"abstain_vote":         proposal.FinalTallyResult.AbstainCount,
		"voting_period_end":    common.TimeToNano(&proposal.VotingPeriodEnd),
		"executor_result":      proposal.ExecutorResult.String(),
		"messages":             messages,
		"title":                proposal.Title,
		"summary":              proposal.Summary,
	})
}

func (h *Hook) doAddGroupMembers(ctx sdk.Context, groupId uint64) {
	groupMembersResponse, _ := h.groupKeeper.GroupMembers(
		sdk.WrapSDKContext(ctx),
		&group.QueryGroupMembersRequest{GroupId: groupId},
	)
	for {
		groupMembers := groupMembersResponse.Members
		for _, groupMember := range groupMembers {
			h.Write("NEW_GROUP_MEMBER", common.JsDict{
				"group_id": groupId,
				"address":  groupMember.Member.Address,
				"weight":   groupMember.Member.Weight,
				"metadata": groupMember.Member.Metadata,
				"added_at": common.TimeToNano(&groupMember.Member.AddedAt),
			})
		}
		if len(groupMembersResponse.Pagination.NextKey) == 0 {
			break
		}
		groupMembersResponse, _ = h.groupKeeper.GroupMembers(
			sdk.WrapSDKContext(ctx),
			&group.QueryGroupMembersRequest{
				GroupId: groupId,
				Pagination: &query.PageRequest{
					Key: groupMembersResponse.Pagination.NextKey,
				},
			},
		)
	}
}
