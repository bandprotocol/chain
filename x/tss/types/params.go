package types

import (
	"fmt"
	"time"
)

const (
	DefaultMaxDESize                             = uint64(100)
	DefaultCreatingPeriod                        = uint64(30000)
	DefaultSigningPeriod                         = uint64(100)
	DefaultInactivePenaltyDuration time.Duration = time.Minute * 10    // 10 minutes
	DefaultJailPenaltyDuration     time.Duration = time.Hour * 24 * 30 // 30 days
)

// NewParams creates a new Params instance
func NewParams(
	maxDESize uint64,
	creatingPeriod uint64,
	signingPeriod uint64,
	inactivePenaltyDuration time.Duration,
	jailPenaltyDuration time.Duration,
) Params {
	return Params{
		MaxDESize:               maxDESize,
		CreatingPeriod:          creatingPeriod,
		SigningPeriod:           signingPeriod,
		InactivePenaltyDuration: inactivePenaltyDuration,
		JailPenaltyDuration:     jailPenaltyDuration,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxDESize:               DefaultMaxDESize,
		CreatingPeriod:          DefaultCreatingPeriod,
		SigningPeriod:           DefaultSigningPeriod,
		InactivePenaltyDuration: DefaultInactivePenaltyDuration,
		JailPenaltyDuration:     DefaultJailPenaltyDuration,
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

	if err := validateTimeDuration("inactive penalty duration")(p.InactivePenaltyDuration); err != nil {
		return err
	}

	if err := validateTimeDuration("jail penalty duration")(p.JailPenaltyDuration); err != nil {
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
