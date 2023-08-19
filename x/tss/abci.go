package tss

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// handleBeginBlock handles the logic at the beginning of a block.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	// Reward a portion of block rewards (inflation + tx fee) to active tss validators.
	k.AllocateTokens(ctx, req.LastCommitInfo.GetVotes())
}

// handleEndBlock handles tasks at the end of a block.
func handleEndBlock(ctx sdk.Context, k *keeper.Keeper) {
	// Get the list of pending process groups.
	gids := k.GetPendingProcessGroups(ctx)
	for _, gid := range gids {
		// Handle the processing for the current pending process group.
		k.HandleProcessGroup(ctx, gid)
	}
	// After processing all pending process groups, set the list of pending process groups to an empty list.
	// This effectively clears the list, as the processing for all groups has been completed in this block.
	k.SetPendingProcessGroups(ctx, types.PendingProcessGroups{})

	// Get the list of pending process signings.
	sids := k.GetPendingProcessSignings(ctx)
	for _, sid := range sids {
		// Handle the processing for the current pending process signing.
		k.HandleProcessSigning(ctx, sid)
	}
	// After processing all pending process signings, set the list of pending process signings to an empty list.
	// This effectively clears the list, as the processing for all signings has been completed in this block.
	k.SetPendingProcessSignings(ctx, types.PendingProcessSignings{})

	// Get the list of pending replace groups.
	pgs := k.GetPendingReplaceGroups(ctx)
	// Create a slice to hold pending replace groups that are not yet processed.
	var pgsLeft []types.PendingReplaceGroup
	for _, prg := range pgs {
		// Check if the execution time is in the future.
		if ctx.BlockTime().Before(prg.ExecTime) {
			// Keep this replace group in the pending list for future processing.
			pgsLeft = append(pgsLeft, prg)
			continue
		}
		// Handle the processing for the current pending replace group.
		k.HandleReplaceGroup(ctx, prg)
	}
	// Set the list of pending replace groups to those that were not processed yet.
	k.SetPendingReplaceGroups(ctx, types.PendingReplaceGroups{PendingReplaceGroups: pgsLeft})

	// Handles cleanup and actions that are required for groups that have expired.
	k.HandleExpiredGroups(ctx)

	// Handles cleanup and actions that are required for signings that have expired.
	k.HandleExpiredSignings(ctx)

	// Handles marking validator as inactive if the validator is not active recently.
	k.HandleInactiveValidators(ctx)
}
