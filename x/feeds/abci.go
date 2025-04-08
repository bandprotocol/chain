package feeds

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// EndBlocker is a handler function for the EndBlock ABCI request.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyEndBlocker)

	// re-calculate current feeds every `CurrentFeedsUpdateInterval` blocks
	if ctx.BlockHeight()%k.GetParams(ctx).CurrentFeedsUpdateInterval == 0 {
		// delete all prices to reset the state
		// it will be set again when the price is calculated in this endblock.
		k.DeleteAllPrices(ctx)

		// update current feeds
		feeds := k.CalculateNewCurrentFeeds(ctx)
		k.SetCurrentFeeds(ctx, feeds)
	}

	// re-calculate prices of all current feeds
	return k.CalculatePrices(ctx)
}
