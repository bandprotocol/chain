package price

import (
	"github.com/bandprotocol/chain/v2/pkg/obi"
)

const DefaultMultiplier = uint64(100000000)

type CommonOutput struct {
	Symbols    []string
	Rates      []uint64
	Multiplier uint64
}

type LegacyInput struct {
	Symbols    []string `json:"symbols"`
	Multiplier uint64   `json:"multiplier"`
}

type LegacyOutput struct {
	Rates []uint64 `json:"rates"`
}

type Input struct {
	Symbols            []string
	MinimumSourceCount uint8
}

type Output struct {
	Responses []Responses
}

type Responses struct {
	Symbol       string
	ResponseCode uint8
	Rate         uint64
}

func MustDecodeResult(calldata, result []byte) CommonOutput {
	var symbols []string
	var rates []uint64

	responses, err := DecodeResult(result)
	if err == nil {
		for _, r := range responses {
			if r.ResponseCode != 0 {
				continue
			}

			symbols = append(symbols, r.Symbol)
			rates = append(rates, r.Rate)
		}

		return CommonOutput{
			Symbols:    symbols,
			Rates:      rates,
			Multiplier: DefaultMultiplier,
		}
	}

	legacyInput, legacyOutput, err := DecodeLegacyResult(calldata, result)
	if err != nil {
		panic(err)
	}

	return CommonOutput{
		Symbols:    legacyInput.Symbols,
		Rates:      legacyOutput.Rates,
		Multiplier: legacyInput.Multiplier,
	}
}

func DecodeLegacyResult(calldata, result []byte) (LegacyInput, LegacyOutput, error) {
	var legacyInput LegacyInput
	var legacyOutput LegacyOutput

	err := obi.Decode(calldata, &legacyInput)
	if err != nil {
		return LegacyInput{}, LegacyOutput{}, err
	}

	err = obi.Decode(result, &legacyOutput)
	if err != nil {
		return LegacyInput{}, LegacyOutput{}, err
	}

	return legacyInput, legacyOutput, nil
}

func DecodeResult(result []byte) ([]Responses, error) {
	var out Output
	err := obi.Decode(result, &out)

	if err != nil {
		return nil, err
	}
	return out.Responses, nil
}
