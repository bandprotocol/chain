package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/keeper"
)

// handleEndBlock cleans up the state during end block. See comment in the implementation!
func handleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	// Process expired groups
	k.ProcessExpiredGroups(ctx)

	// Process expired signings
	k.ProcessExpiredSignings(ctx)
}
