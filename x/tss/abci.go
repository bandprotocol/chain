package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/keeper"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// EndBlocker handles tasks at the end of a block.
func EndBlocker(ctx sdk.Context, k *keeper.Keeper) error {
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

	return nil
}
