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

// RelayPrice represents the price data for relaying to other chains.
type RelayPrice struct {
	SignalID [32]byte
	Price    uint64
}

// NewRelayPrice creates a new RelayPrice instance
func NewRelayPrice(signalID [32]byte, price uint64) RelayPrice {
	return RelayPrice{SignalID: signalID, Price: price}
}

// ToRelayPrices converts a list of prices to RelayPrice
func ToRelayPrices(prices []Price) ([]RelayPrice, error) {
	relayPrices := make([]RelayPrice, 0, len(prices))

	for _, price := range prices {
		signalID, err := StringToBytes32(price.SignalID)
		if err != nil {
			return nil, ErrInvalidSignal.Wrapf("invalid signal id %s: %s", price.SignalID, err)
		}

		relayPrices = append(relayPrices, NewRelayPrice(signalID, price.Price))
	}

	return relayPrices, nil
}

// ToRelayTickPrices converts a list of prices to RelayPrice with price converted to tick
func ToRelayTickPrices(prices []Price) ([]RelayPrice, error) {
	relayPrices := make([]RelayPrice, 0, len(prices))

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

		relayPrices = append(relayPrices, NewRelayPrice(signalID, p))
	}

	return relayPrices, nil
}

// EncodeTSS encodes the feed prices to TSS message
func EncodeTSS(prices []Price, timestamp int64, encoder Encoder) ([]byte, error) {
	switch encoder {
	case ENCODER_FIXED_POINT_ABI:
		relayPrices, err := ToRelayPrices(prices)
		if err != nil {
			return nil, err
		}

		bz, err := feedsPriceDataArgs.Pack(relayPrices, timestamp)
		if err != nil {
			return nil, ErrEncodingPriceFailed.Wrapf("failed to encode price data: %s", err)
		}

		return append([]byte(EncoderFixedPointABIPrefix), bz...), nil
	case ENCODER_TICK_ABI:
		relayPrices, err := ToRelayTickPrices(prices)
		if err != nil {
			return nil, err
		}

		bz, err := feedsPriceDataArgs.Pack(relayPrices, timestamp)
		if err != nil {
			return nil, ErrEncodingPriceFailed.Wrapf("failed to encode price data: %s", err)
		}

		return append([]byte(EncoderTickABIPrefix), bz...), nil
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder: %s", encoder)
	}
}
