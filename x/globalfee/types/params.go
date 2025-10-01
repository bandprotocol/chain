package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewParams returns Params instance with the given values.
func NewParams(minimumGasPrices sdk.DecCoins) Params {
	return Params{
		MinimumGasPrices: minimumGasPrices,
	}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{MinimumGasPrices: sdk.DecCoins{}}
}

// validateMinimumGasPrices checks that the minimum gas prices are non-negative
func validateMinimumGasPrices(i interface{}) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return sdkerrors.ErrInvalidType.Wrapf("type: %T, expected sdk.DecCoins", i)
	}

	return v.Validate()
}

func (p Params) Validate() error {
	return validateMinimumGasPrices(p.MinimumGasPrices)
}
