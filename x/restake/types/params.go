package types

import (
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params instance
func NewParams(
	allowedDenoms []string,
) Params {
	return Params{
		AllowedDenoms: allowedDenoms,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(nil)
}

// Validate validates the set of params
func (p Params) Validate() error {
	for _, denom := range p.AllowedDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return err
		}
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
