package keeper

import (
	"math"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// CreatePricesMap creates a map of prices with signal ID as the key
func CreatePricesMap(prices []feedstypes.Price) map[string]feedstypes.Price {
	pricesMap := make(map[string]feedstypes.Price, len(prices))
	for _, p := range prices {
		pricesMap[p.SignalID] = p
	}
	return pricesMap
}

// GenerateNewPrices generates new prices based on the current prices and signal deviations.
func GenerateNewPrices(
	signalDeviations []types.SignalDeviation,
	latestPricesMap map[string]feedstypes.Price,
	feedsPricesMap map[string]feedstypes.Price,
	timestamp int64,
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
			feedPrice = feedstypes.NewPrice(feedstypes.PRICE_STATUS_NOT_IN_CURRENT_FEEDS, sd.SignalID, 0, timestamp)
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

// calculateDeviationBPS calculates the deviation between the old price and
// the new price in basis points, i.e., |(newPrice - oldPrice)| * 10000 / oldPrice
func calculateDeviationBPS(oldPrice, newPrice sdkmath.Int) sdkmath.Int {
	if newPrice.Equal(oldPrice) {
		return sdkmath.ZeroInt()
	}

	if oldPrice.IsZero() {
		return sdkmath.NewInt(math.MaxInt64)
	}

	return newPrice.Sub(oldPrice).Abs().MulRaw(10000).Quo(oldPrice)
}

// IsOutOfGasError checks if the error object is an out of gas or gas overflow error type
func IsOutOfGasError(err any) (bool, string) {
	switch e := err.(type) {
	case storetypes.ErrorOutOfGas:
		return true, e.Descriptor
	case storetypes.ErrorGasOverflow:
		return true, e.Descriptor
	default:
		return false, ""
	}
}
