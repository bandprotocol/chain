package rollingseed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/rollingseed/keeper"
)

// BeginBlocker re-calculates and saves the rolling seed value based on block hashes.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// Update rolling seed used for pseudorandom oracle provider selection.
	hash := ctx.HeaderInfo().Hash

	// On the first block in the test. it's possible to have empty hash.
	if len(hash) > 0 {
		rollingSeed := k.GetRollingSeed(ctx)
		k.SetRollingSeed(ctx, append(rollingSeed[1:], hash[0]))
	}

	return nil
}
