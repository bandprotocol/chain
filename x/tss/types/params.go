package types

import (
	fmt "fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const DefaultMaxGroupSize = uint64(10)

var KeyMaxGroupSize = []byte("MaxGroupSize")

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxGroupSize uint64) Params {
	return Params{
		MaxGroupSize: maxGroupSize,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{MaxGroupSize: DefaultMaxGroupSize}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMaxGroupSize, &p.MaxGroupSize, validateUint64("max group size", true),
		),
	}
}

func validateUint64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
