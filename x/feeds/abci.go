package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
)

// HandleEndBlock is a handler function for the EndBlock ABCI request.
func HandleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	k.CalculatePrices(ctx)
	if ctx.BlockHeight()%k.GetParams(ctx).CurrentFeedsUpdateInterval == 0 {
		feeds := k.CalculateNewCurrentFeeds(ctx)
		k.SetCurrentFeeds(ctx, feeds)
	}
}
