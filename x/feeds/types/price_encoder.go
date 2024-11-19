package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/bandprotocol/chain/v3/pkg/tickmath"
)

var _priceABI, _ = abi.NewType("tuple[]", "struct Prices[]", []abi.ArgumentMarshaling{
	{Name: "SignalID", Type: "bytes32"},
	{Name: "Price", Type: "uint64"},
})

var _uint64ABI, _ = abi.NewType("uint64", "", nil)

var feedsPriceDataArgs = abi.Arguments{
	abi.Argument{Type: _priceABI, Name: "Prices"},
	abi.Argument{Type: _uint64ABI, Name: "timestamp"},
}

type PriceEncoder struct {
	SignalID [32]byte
	Price    uint64
}

func ToPriceEncoder(price Price, encoder Encoder) (PriceEncoder, error) {
	signalID, err := StringToBytes32(price.SignalID)
	if err != nil {
		return PriceEncoder{}, ErrInvalidSignal.Wrapf("invalid signal id %s: %s", price.SignalID, err)
	}

	if encoder == ENCODER_UNSPECIFIED {
		return PriceEncoder{}, ErrInvalidEncoder.Wrap("encoder is not specified")
	}

	p := price.Price
	if price.Price != 0 && encoder == ENCODER_TICK_ABI {
		tick, err := tickmath.PriceToTick(price.Price)
		if err != nil {
			return PriceEncoder{}, ErrEncodingPriceFailed.Wrapf("failed to convert price to tick: %s", err)
		}

		p = tick
	}

	return PriceEncoder{SignalID: signalID, Price: p}, nil
}

type PriceEncoders []PriceEncoder

func ToPriceEncoders(prices []Price, encoder Encoder) (PriceEncoders, error) {
	var priceEncoders PriceEncoders
	for _, p := range prices {
		pe, err := ToPriceEncoder(p, encoder)
		if err != nil {
			return nil, err
		}

		priceEncoders = append(priceEncoders, pe)
	}

	return priceEncoders, nil
}

func (ps *PriceEncoders) EncodeABI(timestamp uint64) ([]byte, error) {
	bz, err := feedsPriceDataArgs.Pack(ps, timestamp)
	if err != nil {
		return nil, ErrEncodingPriceFailed.Wrapf("failed to encode price data: %s", err)
	}

	return bz, nil
}
