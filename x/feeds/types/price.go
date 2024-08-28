package types

// NewPrice creates a new price instance
func NewPrice(
	priceStatus PriceStatus,
	signalID string,
	price uint64,
	timestamp int64,
) Price {
	return Price{
		PriceStatus: priceStatus,
		SignalID:    signalID,
		Price:       price,
		Timestamp:   timestamp,
	}
}
