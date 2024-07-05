package feeds

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
)

// HandleEndBlock is a handler function for the EndBlock ABCI request.
func HandleEndBlock(ctx sdk.Context, k keeper.Keeper) {
	k.CalculatePrices(ctx)
	if ctx.BlockHeight()%int64(k.GetParams(ctx).SupportedFeedsUpdateInterval) == 0 {
		feeds := k.CalculateNewSupportedFeeds(ctx)
		k.SetSupportedFeeds(ctx, feeds)
	}
}
