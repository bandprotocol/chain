package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/GeoDB-Limited/odin-core/x/oracle/keeper"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

// handleBeginBlock re-calculates and saves the rolling seed value based on block hashes.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k oraclekeeper.Keeper) {
	// Update rolling seed used for pseudorandom oracle provider selection.
	rollingSeed := k.GetRollingSeed(ctx)
	k.SetRollingSeed(ctx, append(rollingSeed[1:], req.GetHash()[0]))
	// Reward a portion of block rewards (inflation + tx fee) to active oracle validators.
	k.AllocateTokens(ctx, req.LastCommitInfo.GetVotes())
}

// handleEndBlock cleans up the state during end block. See comment in the implementation!
func handleEndBlock(ctx sdk.Context, k oraclekeeper.Keeper) {
	// Loops through all requests in the resolvable list to resolve all of them!
	for _, reqID := range k.GetPendingResolveList(ctx) {
		k.ResolveRequest(ctx, reqID)
		k.AllocateRewardsToDataProviders(ctx, reqID)
	}

	rewardThresholdBlocks := k.GetDataProviderRewardThresholdParam(ctx).Blocks
	blockHeight := uint64(ctx.BlockHeight())
	if blockHeight%rewardThresholdBlocks == 0 {
		k.SetAccumulatedDataProvidersRewards(ctx, oracletypes.NewDataProvidersAccumulatedRewards(sdk.NewCoins()))
	}

	// Once all the requests are resolved, we can clear the list.
	k.SetPendingResolveList(ctx, []oracletypes.RequestID{})
	// Lastly, we clean up data requests that are supposed to be expired.
	k.ProcessExpiredRequests(ctx)
	// NOTE: We can remove old requests from state to optimize space, using `k.DeleteRequest`
	// and `k.DeleteReports`. We don't do that now as it is premature optimization at this state.
}
