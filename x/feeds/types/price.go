package types

// NewPrice creates a new price instance
func NewPrice(
	status PriceStatus,
	signalID string,
	price uint64,
	timestamp int64,
) Price {
	return Price{
		Status:    status,
		SignalID:  signalID,
		Price:     price,
		Timestamp: timestamp,
	}
}

// NewSignalPrice creates a new signal price instance
func NewSignalPrice(
	status SignalPriceStatus,
	signalID string,
	price uint64,
) SignalPrice {
	return SignalPrice{
		Status:   status,
		SignalID: signalID,
		Price:    price,
	}
}
