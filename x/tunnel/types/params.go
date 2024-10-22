package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultMinInterval   = uint64(1)
	DefaultMinDeposit    = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	DefaultMaxSignals    = uint64(100)
	DefaultBasePacketFee = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
)

// NewParams creates a new Params instance
func NewParams(
	minDeposit sdk.Coins,
	minInterval uint64,
	maxSignals uint64,
	basePacketFee sdk.Coins,
) Params {
	return Params{
		MinDeposit:    minDeposit,
		MinInterval:   minInterval,
		MaxSignals:    maxSignals,
		BasePacketFee: basePacketFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinDeposit,
		DefaultMinInterval,
		DefaultMaxSignals,
		DefaultBasePacketFee,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	// validate MinDeposit
	if !p.MinDeposit.IsValid() {
		return fmt.Errorf("invalid minimum deposit: %s", p.MinDeposit)
	}

	// validate MinInterval
	if err := validateUint64("min interval", true)(p.MinInterval); err != nil {
		return err
	}

	// validate MaxSignals
	if err := validateUint64("max signals", true)(p.MaxSignals); err != nil {
		return err
	}

	// validate BasePacketFee
	if !p.BasePacketFee.IsValid() {
		return fmt.Errorf("invalid base packet fee: %s", p.BasePacketFee)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateUint64 validates if a given number is a valid uint64.
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
