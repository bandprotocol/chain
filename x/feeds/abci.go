package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
)

// EndBlocker is a handler function for the EndBlock ABCI request.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// re-calculate prices of all current feeds
	k.CalculatePrices(ctx)

	// re-calculate current feeds every `CurrentFeedsUpdateInterval` blocks
	if ctx.BlockHeight()%k.GetParams(ctx).CurrentFeedsUpdateInterval == 0 {
		feeds := k.CalculateNewCurrentFeeds(ctx)
		k.SetCurrentFeeds(ctx, feeds)
	}
	return nil
}
