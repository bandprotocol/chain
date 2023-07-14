package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultMaxGroupSize                 = uint64(20)
	DefaultMaxDESize                    = uint64(100)
	DefaultCreationPeriod time.Duration = time.Minute * 5
	DefaultSigningPeriod  time.Duration = time.Minute * 5
)

var (
	KeyMaxGroupSize   = []byte("MaxGroupSize")
	KeyMaxDESize      = []byte("MaxDESize")
	KeyCreationPeriod = []byte("CreationPeriod")
	KeySigningPeriod  = []byte("SigningPeriod")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxGroupSize uint64,
	maxDESize uint64,
	CreationPeriod time.Duration,
	signingPeriod time.Duration,
) Params {
	return Params{
		MaxGroupSize:   maxGroupSize,
		MaxDESize:      maxDESize,
		CreationPeriod: CreationPeriod,
		SigningPeriod:  signingPeriod,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxGroupSize:   DefaultMaxGroupSize,
		MaxDESize:      DefaultMaxDESize,
		CreationPeriod: DefaultCreationPeriod,
		SigningPeriod:  DefaultSigningPeriod,
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
		paramtypes.NewParamSetPair(KeyCreationPeriod, &p.CreationPeriod, validateTimeDuration("round period")),
		paramtypes.NewParamSetPair(KeySigningPeriod, &p.SigningPeriod, validateTimeDuration("signing period")),
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

func validateTimeDuration(name string) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(time.Duration)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}

		if v.Seconds() <= 0 {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
