package types

import (
	"gopkg.in/yaml.v2"
)

// NewParams creates a new Params instance
func NewParams(
	admin string,
	allowDiffTime int64,
	transitionTime int64,
	minInterval int64,
	maxInterval int64,
	powerThreshold int64,
	maxSupportedSymbol int64,
) Params {
	return Params{
		Admin:              admin,
		AllowDiffTime:      allowDiffTime,
		TransitionTime:     transitionTime,
		MinInterval:        minInterval,
		MaxInterval:        maxInterval,
		PowerThreshold:     powerThreshold,
		MaxSupportedSymbol: maxSupportedSymbol,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams("[NOT_SET]", 30, 30, 60, 3600, 1000_000_000, 5)
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
