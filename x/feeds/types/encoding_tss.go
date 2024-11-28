package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/bandprotocol/chain/v3/pkg/tickmath"
)

const (
	EncoderFixedPointABIPrefix = "\xcb\xa0\xad\x5a" // tss.Hash([]byte("FixedPointABI"))[:4]
	EncoderTickABIPrefix       = "\xdb\x99\xb2\xb3" // tss.Hash([]byte("TickABI"))[:4]
)

var (
	_priceABI, _ = abi.NewType("tuple[]", "struct Prices[]", []abi.ArgumentMarshaling{
		{Name: "SignalID", Type: "bytes32"},
		{Name: "Price", Type: "uint64"},
	})

	_int64ABI, _ = abi.NewType("int64", "", nil)

	feedsPriceDataArgs = abi.Arguments{
		abi.Argument{Type: _priceABI, Name: "Prices"},
		abi.Argument{Type: _int64ABI, Name: "Timestamp"},
	}
)

// TSSPrice represents the price data to be encoded for encoding abi
type TSSPrice struct {
	SignalID [32]byte
	Price    uint64
}

// NewTSSPrice creates a new EncodingPrice instance
func NewTSSPrice(signalID [32]byte, price uint64) TSSPrice {
	return TSSPrice{SignalID: signalID, Price: price}
}

// ToTSSPrices converts a list of prices to TSSPrice
func ToTSSPrices(prices []Price) ([]TSSPrice, error) {
	tssPrices := make([]TSSPrice, 0, len(prices))

	for _, price := range prices {
		signalID, err := StringToBytes32(price.SignalID)
		if err != nil {
			return nil, ErrInvalidSignal.Wrapf("invalid signal id %s: %s", price.SignalID, err)
		}

		tssPrices = append(tssPrices, NewTSSPrice(signalID, price.Price))
	}

	return tssPrices, nil
}

// ToTSSTickPrices converts a list of prices to TSSPrice with price converted to tick
func ToTSSTickPrices(prices []Price) ([]TSSPrice, error) {
	tssPrices := make([]TSSPrice, 0, len(prices))

	for _, price := range prices {
		signalID, err := StringToBytes32(price.SignalID)
		if err != nil {
			return nil, ErrInvalidSignal.Wrapf("invalid signal id %s: %s", price.SignalID, err)
		}

		p := price.Price
		if p != 0 {
			p, err = tickmath.PriceToTick(price.Price)
			if err != nil {
				return nil, ErrEncodingPriceFailed.Wrapf("failed to convert price to tick: %s", err)
			}
		}

		tssPrices = append(tssPrices, NewTSSPrice(signalID, p))
	}

	return tssPrices, nil
}

// EncodeTSS encodes the feed prices to tss message
func EncodeTSS(prices []Price, timestamp int64, encoder Encoder) ([]byte, error) {
	switch encoder {
	case ENCODER_FIXED_POINT_ABI:
		tssPrices, err := ToTSSPrices(prices)
		if err != nil {
			return nil, err
		}

		bz, err := feedsPriceDataArgs.Pack(tssPrices, timestamp)
		if err != nil {
			return nil, ErrEncodingPriceFailed.Wrapf("failed to encode price data: %s", err)
		}

		return append([]byte(EncoderFixedPointABIPrefix), bz...), nil
	case ENCODER_TICK_ABI:
		tssTickPrices, err := ToTSSTickPrices(prices)
		if err != nil {
			return nil, err
		}

		bz, err := feedsPriceDataArgs.Pack(tssTickPrices, timestamp)
		if err != nil {
			return nil, ErrEncodingPriceFailed.Wrapf("failed to encode price data: %s", err)
		}

		return append([]byte(EncoderTickABIPrefix), bz...), nil
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder: %s", encoder)
	}
}
