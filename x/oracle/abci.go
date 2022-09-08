package oracle

import (
	"fmt"

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

func reverseArray(arr []types.RequestID) []types.RequestID {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// handleEndBlock cleans up the state during end block. See comment in the implementation!
func handleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	// jobc := make(chan struct{}, MAX_CONCURRENT_JOBS)

	// var wg sync.WaitGroup

	// Loops through all requests in the resolvable list to parallel resolve all of them!
	rPendingResolveList := reverseArray(k.GetPendingResolveList(ctx))

	fmt.Printf("\n\n\n\nlen pendingResolve: %d \n\n\n\n", len(rPendingResolveList))
	for _, reqID := range rPendingResolveList {
		fmt.Println("pending resolve list")
		k.ResolveRequest(ctx, reqID)
	}

	// for _, reqID := range k.GetPendingResolveList(ctx) {
	// 	wg.Add(1)
	// 	go func(ctx sdk.Context, reqID types.RequestID) {
	// 		defer wg.Done()

	// 		// Create an empty struct to signal when the job finishes
	// 		jobc <- struct{}{}
	// 		k.ResolveRequest(ctx, reqID)
	// 		<-jobc
	// 	}(ctx, reqID)
	// }
	// wg.Wait()

	// Once all the requests are resolved, we can clear the list.
	k.SetPendingResolveList(ctx, []types.RequestID{})
	// Lastly, we clean up data requests that are supposed to be expired.
	k.ProcessExpiredRequests(ctx)
	// NOTE: We can remove old requests from state to optimize space, using `k.DeleteRequest`
	// and `k.DeleteReports`. We don't do that now as it is premature optimization at this state.
}
