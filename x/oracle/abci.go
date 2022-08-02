package oracle

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

const MAX_CONCURRENT_JOBS = 5

// handleBeginBlock re-calculates and saves the rolling seed value based on block hashes.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	// Update rolling seed used for pseudorandom oracle provider selection.
	rollingSeed := k.GetRollingSeed(ctx)
	k.SetRollingSeed(ctx, append(rollingSeed[1:], req.GetHash()[0]))
	// Reward a portion of block rewards (inflation + tx fee) to active oracle validators.
	k.AllocateTokens(ctx, req.LastCommitInfo.GetVotes())
}

// handleEndBlock cleans up the state during end block. See comment in the implementation!
func handleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	jobc := make(chan struct{}, MAX_CONCURRENT_JOBS)

	var wg sync.WaitGroup

	// Loops through all requests in the resolvable list to parallel resolve all of them!
	for _, reqID := range k.GetPendingResolveList(ctx) {
		wg.Add(1)
		// Create an empty struct to signal when the job finishes
		jobc <- struct{}{}
		go k.ResolveRequest(ctx, reqID, jobc)
	}
	wg.Wait()

	// Once all the requests are resolved, we can clear the list.
	k.SetPendingResolveList(ctx, []types.RequestID{})
	// Lastly, we clean up data requests that are supposed to be expired.
	k.ProcessExpiredRequests(ctx)
	// NOTE: We can remove old requests from state to optimize space, using `k.DeleteRequest`
	// and `k.DeleteReports`. We don't do that now as it is premature optimization at this state.
}
