package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var signalPrices, _ = abi.NewType("tuple[]", "struct SignalPrices[]", []abi.ArgumentMarshaling{
	{Name: "SignalID", Type: "bytes32"},
	{Name: "Price", Type: "uint64"},
})

var _uint64, _ = abi.NewType("uint64", "", nil)

var feedsPriceDataArgs = abi.Arguments{
	abi.Argument{Type: signalPrices, Name: "signalPrices"},
	abi.Argument{Type: _uint64, Name: "timestamp"},
}

type SignalPriceABI struct {
	SignalID [32]byte
	Price    uint64
}

func NewFeedsPriceData(signalPrices []SignalPrice, timestamp uint64) *FeedsPriceData {
	return &FeedsPriceData{
		SignalPrices: signalPrices,
		Timestamp:    timestamp,
	}
}

func (f FeedsPriceData) ABIEncode() ([]byte, error) {
	signalPriceABIs := make([]SignalPriceABI, len(f.SignalPrices))

	for i, signalPrice := range f.SignalPrices {
		signalID, err := stringToBytes32(signalPrice.SignalID)
		if err != nil {
			return nil, ErrInvalidSignal.Wrapf(
				"invalid signal id %s: %s", signalPrice.SignalID, err,
			)
		}

		signalPriceABIs[i] = SignalPriceABI{
			SignalID: signalID,
			Price:    signalPrice.Price,
		}
	}

	return feedsPriceDataArgs.Pack(signalPriceABIs, f.Timestamp)
}
