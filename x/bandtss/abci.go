package bandtss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
)

// handleBeginBlock handles the logic at the beginning of a block.
func handleBeginBlock(ctx sdk.Context, k *keeper.Keeper) {
	// Reward a portion of block rewards (inflation + tx fee) to active tss members.
	k.AllocateTokens(ctx)
}

// handleEndBlock handles tasks at the end of a block.
func handleEndBlock(ctx sdk.Context, k *keeper.Keeper) {
	// execute group transition if the transition execution time is reached.
	if transition, ok := k.ShouldExecuteGroupTransition(ctx); ok {
		k.ExecuteGroupTransition(ctx, transition)
	}

	// Handles marking members as inactive if the member is not active recently.
	k.HandleInactiveMembers(ctx)
}
