package keeper

import (
	sdkmath "cosmossdk.io/math"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// GenerateNewSignalPrices generates new signal prices based on the current prices
// and signal deviations.
func GenerateNewSignalPrices(
	latestSignalPrices types.LatestSignalPrices,
	signalDeviationsMap map[string]types.SignalDeviation,
	currentFeedsPricesMap map[string]feedstypes.Price,
	sendAll bool,
) ([]types.SignalPrice, error) {
	shouldSend := false
	newSignalPrices := make([]types.SignalPrice, 0)
	for _, sp := range latestSignalPrices.SignalPrices {
		oldPrice := sdkmath.NewIntFromUint64(sp.Price)

		// get current price from the feed, if not found, set price to 0
		price := uint64(0)
		feedPrice, ok := currentFeedsPricesMap[sp.SignalID]
		if ok && feedPrice.PriceStatus == feedstypes.PriceStatusAvailable {
			price = feedPrice.Price
		}
		newPrice := sdkmath.NewIntFromUint64(price)

		// get hard/soft deviation, panic if not found; should not happen.
		sd, ok := signalDeviationsMap[sp.SignalID]
		if !ok {
			return nil, types.ErrDeviationNotFound.Wrapf("deviation not found for signal ID :%s", sp.SignalID)
		}
		hardDeviation := sdkmath.NewIntFromUint64(sd.HardDeviationBPS)
		softDeviation := sdkmath.NewIntFromUint64(sd.SoftDeviationBPS)

		// calculate deviation between old price and new price and compare with the threshold.
		// shouldSend is set to true if sendAll is true or there is a signal whose deviation
		// is over the hard threshold.
		deviation := calculateDeviationBPS(oldPrice, newPrice)
		if sendAll || deviation.GTE(hardDeviation) {
			newSignalPrices = append(newSignalPrices, types.NewSignalPrice(sp.SignalID, price))
			shouldSend = true
		} else if deviation.GTE(softDeviation) {
			newSignalPrices = append(newSignalPrices, types.NewSignalPrice(sp.SignalID, price))
		}
	}

	if shouldSend {
		return newSignalPrices, nil
	} else {
		return []types.SignalPrice{}, nil
	}
}
