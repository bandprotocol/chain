package keeper

import (
	sdkmath "cosmossdk.io/math"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// GenerateNewPrices generates new prices based on the current prices and signal deviations.
func GenerateNewPrices(
	signalDeviations []types.SignalDeviation,
	latestPricesMap map[string]feedstypes.Price,
	feedsPricesMap map[string]feedstypes.Price,
	sendAll bool,
) []feedstypes.Price {
	shouldSend := false

	newFeedPrices := make([]feedstypes.Price, 0)
	for _, sd := range signalDeviations {
		oldPrice := sdkmath.NewInt(0)
		if latestPrices, ok := latestPricesMap[sd.SignalID]; ok {
			oldPrice = sdkmath.NewIntFromUint64(latestPrices.Price)
		}

		feedPrice, ok := feedsPricesMap[sd.SignalID]
		if !ok {
			feedPrice = feedstypes.NewPrice(feedstypes.PRICE_STATUS_NOT_IN_CURRENT_FEEDS, sd.SignalID, 0, 0)
		}

		// calculate deviation between old price and new price and compare with the threshold.
		// shouldSend is set to true if sendAll is true or there is a signal whose deviation
		// is over the hard threshold.
		deviation := calculateDeviationBPS(oldPrice, sdkmath.NewIntFromUint64(feedPrice.Price))
		if sendAll || deviation.GTE(sdkmath.NewIntFromUint64(sd.HardDeviationBPS)) {
			newFeedPrices = append(newFeedPrices, feedPrice)
			shouldSend = true
		} else if deviation.GTE(sdkmath.NewIntFromUint64(sd.SoftDeviationBPS)) {
			newFeedPrices = append(newFeedPrices, feedPrice)
		}
	}

	if shouldSend {
		return newFeedPrices
	} else {
		return []feedstypes.Price{}
	}
}
