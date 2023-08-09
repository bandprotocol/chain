package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultMaxGroupSize                          = uint64(20)
	DefaultMaxDESize                             = uint64(100)
	DefaultCreatingPeriod                        = int64(30000)
	DefaultSigningPeriod                         = int64(100)
	DefaultActiveDuration          time.Duration = time.Hour * 24      // 1 days
	DefaultInactivePenaltyDuration time.Duration = time.Minute * 10    // 10 minutes
	DefaultJailPenaltyDuration     time.Duration = time.Hour * 24 * 30 // 30 days
)

var (
	KeyMaxGroupSize            = []byte("MaxGroupSize")
	KeyMaxDESize               = []byte("MaxDESize")
	KeyCreatingPeriod          = []byte("CreatingPeriod")
	KeySigningPeriod           = []byte("SigningPeriod")
	KeyActiveDuration          = []byte("ActiveDuration")
	KeyInactivePenaltyDuration = []byte("InactivePenaltyDuration")
	KeyJailPenaltyDuration     = []byte("JailPenaltyDuration")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxGroupSize uint64,
	maxDESize uint64,
	creatingPeriod int64,
	signingPeriod int64,
	activeDuration time.Duration,
	inactivePenaltyDuration time.Duration,
	jailPenaltyDuration time.Duration,
) Params {
	return Params{
		MaxGroupSize:            maxGroupSize,
		MaxDESize:               maxDESize,
		CreatingPeriod:          creatingPeriod,
		SigningPeriod:           signingPeriod,
		ActiveDuration:          activeDuration,
		InactivePenaltyDuration: inactivePenaltyDuration,
		JailPenaltyDuration:     jailPenaltyDuration,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MaxGroupSize:            DefaultMaxGroupSize,
		MaxDESize:               DefaultMaxDESize,
		CreatingPeriod:          DefaultCreatingPeriod,
		SigningPeriod:           DefaultSigningPeriod,
		ActiveDuration:          DefaultActiveDuration,
		InactivePenaltyDuration: DefaultInactivePenaltyDuration,
		JailPenaltyDuration:     DefaultJailPenaltyDuration,
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
			KeyCreatingPeriod,
			&p.CreatingPeriod,
			validateInt64("creating period", true),
		),
		paramtypes.NewParamSetPair(KeySigningPeriod, &p.SigningPeriod, validateInt64("signing period", true)),
		paramtypes.NewParamSetPair(
			KeyActiveDuration,
			&p.ActiveDuration,
			validateTimeDuration("active duration"),
		),
		paramtypes.NewParamSetPair(
			KeyInactivePenaltyDuration,
			&p.InactivePenaltyDuration,
			validateTimeDuration("inactive penalty duration"),
		),
		paramtypes.NewParamSetPair(
			KeyJailPenaltyDuration,
			&p.JailPenaltyDuration,
			validateTimeDuration("jail penalty duration"),
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
