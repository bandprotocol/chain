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

	switch encoder {
	case types.ENCODER_TICK_ABI:
		var tickPrices []types.Price
		for i := range prices {
			tickPrice, err := prices[i].ToTick()
			if err != nil {
				return nil, err
			}

			tickPrices = append(tickPrices, tickPrice)
		}

		return types.NewFeedsPriceData(tickPrices, uint64(ctx.BlockTime().Unix())), nil
	case types.ENCODER_FIXED_POINT_ABI:
		return types.NewFeedsPriceData(prices, uint64(ctx.BlockTime().Unix())), nil
	default:
		return nil, types.ErrInvalidEncoder
	}
}
