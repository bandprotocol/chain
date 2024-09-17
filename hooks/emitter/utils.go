package emitter

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ConvertToGas(owasm uint64) uint64 {
	// TODO: Using `gasConversionFactor` from oracle module
	return uint64(math.Ceil(float64(owasm) / float64(20_000_000)))
}

func MustParseValAddress(addr string) sdk.ValAddress {
	val, err := sdk.ValAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return val
}
