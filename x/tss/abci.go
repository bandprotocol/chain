package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// handleEndBlock handles tasks at the end of a block.
func handleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	// Handles cleanup and actions that are required for groups that have expired.
	k.HandleExpiredGroups(ctx)

	// Handles cleanup and actions that are required for signings that have expired.
	k.HandleExpiredSignings(ctx)

	// Get the list of pending process groups.
	pgs := k.GetPendingProcessGroups(ctx)
	for _, pg := range pgs {
		// Handle the processing for the current pending process group.
		k.HandleProcessGroup(ctx, pg)
	}

	// After processing all pending process groups, set the list of pending process groups to an empty list.
	// This effectively clears the list, as the processing for all groups has been completed in this block.
	k.SetPendingProcessGroups(ctx, types.PendingProcessGroups{})
}
