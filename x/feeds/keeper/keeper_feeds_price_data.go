package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// GetFeedsPriceData returns the price data of the given signalIDs
func (k Keeper) GetFeedsPriceData(
	ctx sdk.Context,
	signalIDs []string,
	encoder types.Encoder,
) (*types.FeedsPriceData, error) {
	feeds := make(map[string]types.Feed)
	sp := k.GetCurrentFeeds(ctx)
	for _, feed := range sp.Feeds {
		feeds[feed.SignalID] = feed
	}

	var prices []types.SignalPrice
	for _, signalID := range signalIDs {
		// Get the price of the signal
		p, err := k.GetPrice(ctx, signalID)
		if err != nil {
			return nil, err
		}

		// Check if the encoder mode is tick
		if encoder == types.ENCODER_TICK_ABI {
			err := p.ToTick()
			if err != nil {
				return nil, err
			}
		}

		// Check if the price is available
		if p.PriceStatus != types.PriceStatusAvailable {
			return nil, fmt.Errorf("%s: price not available", signalID)
		}

		f, ok := feeds[signalID]
		if !ok {
			return nil, fmt.Errorf("%s: feed not supported", signalID)
		}

		// Check if the price is too old
		if ctx.BlockTime().Unix() > p.Timestamp+f.Interval {
			return nil, fmt.Errorf("%s: price too old", signalID)
		}

		// Append the price to the list
		prices = append(prices, types.SignalPrice{
			SignalID: signalID,
			Price:    p.Price,
		})
	}
	return types.NewFeedsPriceData(prices, uint64(ctx.BlockTime().Unix())), nil
}
