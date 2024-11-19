package types

import (
	fmt "fmt"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

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
