package bandtss

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
)

// handleBeginBlock handles the logic at the beginning of a block.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k *keeper.Keeper) {
	// Reward a portion of block rewards (inflation + tx fee) to active tss validators.
	k.AllocateTokens(ctx, req.LastCommitInfo.GetVotes())
}

// handleEndBlock handles tasks at the end of a block.
func handleEndBlock(ctx sdk.Context, k *keeper.Keeper) {
	// Handle replacement of the current group with the new group.
	if err := k.HandleReplaceGroup(ctx, ctx.BlockHeader().Time); err != nil {
		panic(err)
	}

	// Handles marking validator as inactive if the validator is not active recently.
	k.HandleInactiveValidators(ctx)
}
