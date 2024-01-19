package types

import (
	"gopkg.in/yaml.v2"
)

// NewParams creates a new Params instance
func NewParams(admin string, codeStartTime int64) Params {
	return Params{
		Admin: admin,
		// TODO : adjust to be 1 day.
		ColdStartTime: codeStartTime,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams("[NOT_SET]", 120)
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
