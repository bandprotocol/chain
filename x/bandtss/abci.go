package bandtss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/bandtss/keeper"
)

// BeginBlocker handles the logic at the beginning of a block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// Reward a portion of block rewards (inflation + tx fee) to active tss members.
	k.AllocateTokens(ctx)

	return nil
}

// EndBlocker handles tasks at the end of a block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// execute group transition if the transition execution time is reached.
	if transition, ok := k.ShouldExecuteGroupTransition(ctx); ok {
		k.ExecuteGroupTransition(ctx, transition)
	}

	// Handles marking members as inactive if the member is not active recently.
	k.DeactivateInactiveMembers(ctx)

	return nil
}
