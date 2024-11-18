package types

import (
	"fmt"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

// NewSignalDeviation creates a new SignalDeviation instance.
func NewSignalDeviation(
	signalID string,
	softDeviationBPS uint64,
	hardDeviationBPS uint64,
) SignalDeviation {
	return SignalDeviation{
		SignalID:         signalID,
		SoftDeviationBPS: softDeviationBPS,
		HardDeviationBPS: hardDeviationBPS,
	}
}

// NewLatestPrices creates a new LatestPrices instance.
func NewLatestPrices(
	tunnelID uint64,
	prices []feedstypes.Price,
	lastInterval int64,
) LatestPrices {
	return LatestPrices{
		TunnelID:     tunnelID,
		Prices:       prices,
		LastInterval: lastInterval,
	}
}

// Validate validates the latest prices.
func (l LatestPrices) Validate() error {
	if l.TunnelID == 0 {
		return fmt.Errorf("tunnel ID cannot be 0")
	}

	if l.LastInterval < 0 {
		return fmt.Errorf("last interval cannot be negative")
	}

	return nil
}

// UpdatePrices updates prices in the latest prices.
func (l *LatestPrices) UpdatePrices(newPrices []feedstypes.Price) {
	pricesIndex := make(map[string]int)
	for i, p := range l.Prices {
		pricesIndex[p.SignalID] = i
	}

	for _, p := range newPrices {
		if i, ok := pricesIndex[p.SignalID]; ok {
			l.Prices[i] = p
		} else {
			l.Prices = append(l.Prices, p)
			pricesIndex[p.SignalID] = len(l.Prices) - 1
		}
	}
}

// ValidateSignalDeviations validates the signal deviations with the given params.
func ValidateSignalDeviations(signalDeviations []SignalDeviation, params Params) error {
	// validate max signals
	if len(signalDeviations) > int(params.MaxSignals) {
		return ErrMaxSignalsExceeded.Wrapf("max signals %d, got %d", params.MaxSignals, len(signalDeviations))
	}

	// validate min and max deviation
	for _, signalDeviation := range signalDeviations {
		if signalDeviation.HardDeviationBPS < params.MinDeviationBPS ||
			signalDeviation.SoftDeviationBPS < params.MinDeviationBPS ||
			signalDeviation.HardDeviationBPS > params.MaxDeviationBPS ||
			signalDeviation.SoftDeviationBPS > params.MaxDeviationBPS {
			return ErrDeviationOutOfRange.Wrapf(
				"min %d, max %d, got %d, %d",
				params.MinDeviationBPS,
				params.MaxDeviationBPS,
				signalDeviation.SoftDeviationBPS,
				signalDeviation.HardDeviationBPS,
			)
		}
	}
	return nil
}
