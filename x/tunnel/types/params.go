package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Each value below is the default value for each parameter when generating the default
	// genesis file. See comments in types.proto for explanation for each parameter.
	DefaultMinInterval     = uint64(60)
	DefaultMaxInterval     = uint64(3600)
	DefaultMinDeviationBPS = uint64(50)
	DefaultMaxDeviationBPS = uint64(3000)
	DefaultMinDeposit      = sdk.NewCoins(sdk.NewInt64Coin("uband", 1_000_000_000))
	DefaultMaxSignals      = uint64(25)
	DefaultBasePacketFee   = sdk.NewCoins(sdk.NewInt64Coin("uband", 10_000))
)

// NewParams creates a new Params instance
func NewParams(
	minDeposit sdk.Coins,
	minInterval uint64,
	maxInterval uint64,
	minDeviationBPS uint64,
	maxDeviationBPS uint64,
	maxSignals uint64,
	basePacketFee sdk.Coins,
) Params {
	return Params{
		MinDeposit:      minDeposit,
		MinInterval:     minInterval,
		MaxInterval:     maxInterval,
		MinDeviationBPS: minInterval,
		MaxDeviationBPS: maxDeviationBPS,
		MaxSignals:      maxSignals,
		BasePacketFee:   basePacketFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinDeposit,
		DefaultMinInterval,
		DefaultMaxInterval,
		DefaultMinDeviationBPS,
		DefaultMaxDeviationBPS,
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

	// validate MaxInterval
	if err := validateUint64("max interval", true)(p.MaxInterval); err != nil {
		return err
	}

	// validate max interval is greater than min interval
	if p.MaxInterval <= p.MinInterval {
		return fmt.Errorf("max interval must be greater than min interval: %d <= %d", p.MaxInterval, p.MinInterval)
	}

	// validate MinDeviationBPS
	if err := validateUint64("min deviation bps", true)(p.MinDeviationBPS); err != nil {
		return err
	}

	// validate MaxDeviationBPS
	if err := validateUint64("max deviation bps", true)(p.MaxDeviationBPS); err != nil {
		return err
	}

	// validate max deviation bps is greater than min deviation bps
	if p.MaxDeviationBPS <= p.MinDeviationBPS {
		return fmt.Errorf(
			"max deviation bps must be greater than min deviation bps: %d <= %d",
			p.MaxDeviationBPS,
			p.MinDeviationBPS,
		)
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
