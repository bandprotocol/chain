package types

import (
	"fmt"
)

const (
	DefaultMaxDESize      = uint64(100)
	DefaultCreatingPeriod = uint64(30000)
	DefaultSigningPeriod  = uint64(100)
)

// NewParams creates a new Params instance
func NewParams(
	maxDESize uint64,
	creatingPeriod uint64,
	signingPeriod uint64,
) Params {
	return Params{
		MaxDESize:      maxDESize,
		CreatingPeriod: creatingPeriod,
		SigningPeriod:  signingPeriod,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxDESize:      DefaultMaxDESize,
		CreatingPeriod: DefaultCreatingPeriod,
		SigningPeriod:  DefaultSigningPeriod,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateUint64("max DE size", true)(p.MaxDESize); err != nil {
		return err
	}

	if err := validateUint64("creating period", true)(p.CreatingPeriod); err != nil {
		return err
	}

	if err := validateUint64("signing period", true)(p.SigningPeriod); err != nil {
		return err
	}

	return nil
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
