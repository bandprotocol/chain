package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultMaxGroupSize           = uint64(20)
	DefaultMaxDESize              = uint64(100)
	DefaultGroupSigCreatingPeriod = int64(100)
	DefaultSigningPeriod          = int64(100)
)

var (
	KeyMaxGroupSize           = []byte("MaxGroupSize")
	KeyMaxDESize              = []byte("MaxDESize")
	KeyRoundPeriod            = []byte("RoundPeriod")
	KeyGroupSigCreatingPeriod = []byte("GroupSigCreatingPeriod")
	KeySigningPeriod          = []byte("SigningPeriod")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxGroupSize uint64,
	maxDESize uint64,
	groupSigCreatingPeriod int64,
	signingPeriod int64,
) Params {
	return Params{
		MaxGroupSize:           maxGroupSize,
		MaxDESize:              maxDESize,
		GroupSigCreatingPeriod: groupSigCreatingPeriod,
		SigningPeriod:          signingPeriod,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxGroupSize:           DefaultMaxGroupSize,
		MaxDESize:              DefaultMaxDESize,
		GroupSigCreatingPeriod: DefaultGroupSigCreatingPeriod,
		SigningPeriod:          DefaultSigningPeriod,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxGroupSize, &p.MaxGroupSize, validateUint64("max group size", true)),
		paramtypes.NewParamSetPair(KeyMaxDESize, &p.MaxDESize, validateUint64("max DE size", true)),
		paramtypes.NewParamSetPair(
			KeyGroupSigCreatingPeriod,
			&p.GroupSigCreatingPeriod,
			validateInt64("group sig creating period", true),
		),
		paramtypes.NewParamSetPair(KeySigningPeriod, &p.SigningPeriod, validateInt64("signing period", true)),
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

func validateInt64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(int64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
