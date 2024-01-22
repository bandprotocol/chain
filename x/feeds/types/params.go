package types

import (
	"gopkg.in/yaml.v2"
)

// NewParams creates a new Params instance
func NewParams(admin string, codeStartTime int64, allowGapTime int64) Params {
	return Params{
		Admin: admin,
		// TODO : adjust to be 1 day.
		ColdStartTime: codeStartTime,
		AllowGapTime:  allowGapTime,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	// TODO: adjust the default parameters.
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
