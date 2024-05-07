package rollingseed

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/rollingseed/keeper"
)

// handleBeginBlock re-calculates and saves the rolling seed value based on block hashes.
func handleBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper) {
	// Update rolling seed used for pseudorandom oracle provider selection.
	hash := req.GetHash()
	// On the first block in the test. it's possible to have empty hash.
	if len(hash) > 0 {
		rollingSeed := k.GetRollingSeed(ctx)
		k.SetRollingSeed(ctx, append(rollingSeed[1:], hash[0]))
	}
}
