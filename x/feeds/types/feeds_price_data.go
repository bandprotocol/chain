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

func NewFeedsPriceData(prices []Price, timestamp uint64) *FeedsPriceData {
	return &FeedsPriceData{
		Prices:    prices,
		Timestamp: timestamp,
	}
}

func (f FeedsPriceData) ABIEncode() ([]byte, error) {
	signalPriceABIs := make([]SignalPriceABI, len(f.Prices))

	for i, price := range f.Prices {
		signalID, err := stringToBytes32(price.SignalID)
		if err != nil {
			return nil, ErrInvalidSignal.Wrapf(
				"invalid signal id %s: %s", price.SignalID, err,
			)
		}

		signalPriceABIs[i] = SignalPriceABI{
			SignalID: signalID,
			Price:    price.Price,
		}
	}

	return feedsPriceDataArgs.Pack(signalPriceABIs, f.Timestamp)
}

// ValidateEncoder validates the encoder.
func ValidateEncoder(encoder Encoder) error {
	if _, ok := Encoder_name[int32(encoder)]; ok && encoder != ENCODER_UNSPECIFIED {
		return nil
	}

	return ErrInvalidEncoder.Wrapf("invalid encoder: %s", encoder)
}
