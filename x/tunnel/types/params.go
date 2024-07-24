package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"gopkg.in/yaml.v2"
)

var (
	DefaultMinDeposit      = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	DefaultMinDeviationBPS = uint64(100)
	DefaultTSSRouteFee     = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	DefaultAxelarRouteFee  = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
)

// NewParams creates a new Params instance
func NewParams(
	minDeposit sdk.Coins,
	minDeviationBPS uint64,
	tssRouteFee sdk.Coins,
	axelarRouteFee sdk.Coins,
) Params {
	return Params{
		MinDeposit:      minDeposit,
		MinDeviationBPS: minDeviationBPS,
		TSSRouteFee:     tssRouteFee,
		AxelarRouteFee:  axelarRouteFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinDeposit,
		DefaultMinDeviationBPS,
		DefaultTSSRouteFee,
		DefaultAxelarRouteFee,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	// Validate MinDeposit
	if !p.MinDeposit.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf(p.MinDeposit.String())
	}
	// Validate MinDeviationBPS uint64
	if err := validateUint64("min deviation BPS", false)(p.MinDeviationBPS); err != nil {
		return err
	}
	// Validate MinDeviationBPS
	if err := validateBasisPoint(p.MinDeviationBPS); err != nil {
		return err
	}
	// Validate TSSRouteFee
	if !p.TSSRouteFee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf(p.TSSRouteFee.String())
	}
	// Validate AxelarRouteFee
	if !p.AxelarRouteFee.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrapf(p.AxelarRouteFee.String())
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateBasisPoint validates if a given number is a valid basis point (0 to 10000).
func validateBasisPoint(bp uint64) error {
	if bp > 10000 {
		return fmt.Errorf("invalid basis point: must be between 0 and 10000")
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
