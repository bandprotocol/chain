package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetFeedsPriceData returns the price data of the given signalIDs
func (k Keeper) GetFeedsPriceData(
	ctx sdk.Context,
	signalIDs []string,
	ft types.FeedType,
) (*types.FeedsPriceData, error) {
	var prices []types.SignalPrice
	for _, signalID := range signalIDs {
		// Get the price of the signal
		p, err := k.GetPrice(ctx, signalID)
		if err != nil {
			return nil, err
		}

		// Check if the feed type is tick
		if ft == types.FEED_TYPE_TICK {
			err := p.ToTick()
			if err != nil {
				return nil, err
			}
		}

		// Check if the price is available
		if p.PriceStatus != types.PriceStatusAvailable {
			return nil, fmt.Errorf("%s: price not available", signalID)
		}

		// Check if the price is too old
		if ctx.BlockTime().Sub(time.Unix(p.Timestamp, 0)) > types.MAX_PRICE_TIME_DIFF {
			return nil, fmt.Errorf("%s: price too old", signalID)
		}

		// Append the price to the list
		prices = append(prices, types.SignalPrice{
			SignalID: signalID,
			Price:    p.Price,
		})
	}
	return types.NewFeedsPriceData(prices, ctx.BlockTime().Unix()), nil
}
