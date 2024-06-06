package types

import (
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var signalPrices, _ = abi.NewType("tuple[]", "struct SignalPrices[]", []abi.ArgumentMarshaling{
	{Name: "SignalID", Type: "string"},
	{Name: "Price", Type: "uint64"},
})

var _int64, _ = abi.NewType("int64", "", nil)

var feedsPriceDataArgs = abi.Arguments{
	abi.Argument{Type: signalPrices, Name: "signalPrices"},
	abi.Argument{Type: _int64, Name: "timestamp"},
}

// MAX_PRICE_TIME_DIFF is the maximum time difference between the current block time
const MAX_PRICE_TIME_DIFF = time.Second * 10

func NewFeedsPriceData(signalPrices []SignalPrice, timestamp int64) *FeedsPriceData {
	return &FeedsPriceData{
		SignalPrices: signalPrices,
		Timestamp:    timestamp,
	}
}

func (f FeedsPriceData) ABIEncode() ([]byte, error) {
	return feedsPriceDataArgs.Pack(f.SignalPrices, f.Timestamp)
}
