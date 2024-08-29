package tss

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// handleBeginBlock handles the logic at the beginning of a block.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k *keeper.Keeper) {}

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

	// Handles cleanup and actions that are required for groups that have expired.
	k.HandleExpiredGroups(ctx)

	// Handle the signings that should be processed.
	k.HandleSigningEndBlock(ctx)
}
