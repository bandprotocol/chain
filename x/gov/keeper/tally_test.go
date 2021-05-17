package keeper_test

import (
	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTallyNoOneVotes(t *testing.T) {
	app, ctx, _ := testapp.CreateAppCustomBalances(5, 5, 5)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.True(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyNoQuorum(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(2, 5, 0)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	err = app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionYes)
	require.Nil(t, err)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)
	require.False(t, passes)
	require.True(t, burnDeposits)
}

func TestTallyOnlyValidatorsAllYes(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(5, 5, 5)
	tp := TestProposal

	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[2].Address, types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidators51No(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(5, 6, 0)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionNo))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, _ := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
}

func TestTallyOnlyValidators51Yes(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(5, 6, 0)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionNo))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsVetoed(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(6, 6, 7)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[2].Address, types.OptionNoWithVeto))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.True(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainPasses(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(6, 6, 7)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionAbstain))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionNo))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[2].Address, types.OptionYes))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.True(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsAbstainFails(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(6, 6, 7)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionAbstain))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[2].Address, types.OptionNo))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}

func TestTallyOnlyValidatorsNonVoter(t *testing.T) {
	app, ctx, builder := testapp.CreateAppCustomBalances(5, 6, 7)

	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalId
	proposal.Status = types.StatusVotingPeriod
	app.GovKeeper.SetProposal(ctx, proposal)

	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[0].Address, types.OptionYes))
	require.NoError(t, app.GovKeeper.AddVote(ctx, proposalID, builder.GetAuthBuilder().Accounts[1].Address, types.OptionNo))

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	passes, burnDeposits, tallyResults := app.GovKeeper.Tally(ctx, proposal)

	require.False(t, passes)
	require.False(t, burnDeposits)
	require.False(t, tallyResults.Equals(types.EmptyTallyResult()))
}
