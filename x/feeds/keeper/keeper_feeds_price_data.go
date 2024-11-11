package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// GetFeedsPriceData returns the price data of the given signalIDs
func (k Keeper) GetFeedsPriceData(
	ctx sdk.Context,
	signalIDs []string,
	encoder types.Encoder,
) (*types.FeedsPriceData, error) {
	prices := k.GetPrices(ctx, signalIDs)

	// convert price to tick if encoder is tick abi
	if encoder == types.ENCODER_TICK_ABI {
		for i := range prices {
			if err := prices[i].ToTick(); err != nil {
				return nil, err
			}
		}
	}

	return types.NewFeedsPriceData(prices, uint64(ctx.BlockTime().Unix())), nil
}
