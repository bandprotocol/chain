package types

import (
	"gopkg.in/yaml.v2"
)

// NewParams creates a new Params instance
func NewParams(admin string, prepareTime int64, allowDiffTime int64) Params {
	return Params{
		Admin:         admin,
		PrepareTime:   prepareTime,
		AllowDiffTime: allowDiffTime,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	// TODO: adjust the default parameters.
	// - prepare time: 1 day.
	return NewParams("[NOT_SET]", 60, 30)
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
