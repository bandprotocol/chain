package types

import (
	"fmt"
)

const (
	DefaultMaxGroupSize      = uint64(20)
	DefaultMaxDESize         = uint64(100)
	DefaultCreationPeriod    = uint64(30000)
	DefaultSigningPeriod     = uint64(100)
	DefaultMaxSigningAttempt = uint64(5)
	DefaultMaxMemoLength     = uint64(100)
	DefaultMaxMessageLength  = uint64(300)
)

// NewParams creates a new Params instance
func NewParams(
	maxGroupSize uint64,
	maxDESize uint64,
	creatingPeriod uint64,
	signingPeriod uint64,
	maxSigningAttempt uint64,
	maxMemoLength uint64,
	maxMessageLength uint64,
) Params {
	return Params{
		MaxGroupSize:      maxGroupSize,
		MaxDESize:         maxDESize,
		CreationPeriod:    creatingPeriod,
		SigningPeriod:     signingPeriod,
		MaxSigningAttempt: maxSigningAttempt,
		MaxMemoLength:     maxMemoLength,
		MaxMessageLength:  maxMessageLength,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMaxGroupSize,
		DefaultMaxDESize,
		DefaultCreationPeriod,
		DefaultSigningPeriod,
		DefaultMaxSigningAttempt,
		DefaultMaxMemoLength,
		DefaultMaxMessageLength,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	fields := []struct {
		name           string
		val            uint64
		isPositiveOnly bool
	}{
		{"max group size", p.MaxGroupSize, true},
		{"max DE size", p.MaxDESize, true},
		{"creation period", p.CreationPeriod, true},
		{"signing period", p.SigningPeriod, true},
		{"max signing attempt", p.MaxSigningAttempt, false},
		{"max memo length", p.MaxMemoLength, true},
		{"max message length", p.MaxMessageLength, true},
	}

	for _, f := range fields {
		if err := validateUint64(f.name, f.isPositiveOnly)(f.val); err != nil {
			return err
		}
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
