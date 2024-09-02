package types

// NewSignalInfo creates a new SignalInfo instance.
func NewSignalInfo(
	signalID string,
	softDeviationBPS uint64,
	hardDeviationBPS uint64,
) SignalInfo {
	return SignalInfo{
		SignalID:         signalID,
		SoftDeviationBPS: softDeviationBPS,
		HardDeviationBPS: hardDeviationBPS,
	}
}

// NewSignalPricesInfo creates a new SignalPricesInfo instance.
func NewSignalPricesInfo(
	tunnelID uint64,
	signalPrices []SignalPrice,
	lastIntervalTimestamp int64,
) SignalPricesInfo {
	return SignalPricesInfo{
		TunnelID:              tunnelID,
		SignalPrices:          signalPrices,
		LastIntervalTimestamp: lastIntervalTimestamp,
	}
}

// UpdateSignalPrices updates the signal prices based on signal IDs
func (spsi *SignalPricesInfo) UpdateSignalPrices(newSignalPrices []SignalPrice) {
	for _, nsp := range newSignalPrices {
		for i, sp := range spsi.SignalPrices {
			if nsp.SignalID == sp.SignalID {
				spsi.SignalPrices[i] = sp
				break
			}
		}
	}
}

// NewSignalPrice creates a new SignalPrice instance.
func NewSignalPrice(
	signalID string,
	price uint64,
	timestamp int64,
) SignalPrice {
	return SignalPrice{
		SignalID:  signalID,
		Price:     price,
		Timestamp: timestamp,
	}
}
