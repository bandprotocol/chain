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
	timeDividend int64,
	maxSupportedSymbol uint64,
) Params {
	return Params{
		Admin:              admin,
		AllowDiffTime:      allowDiffTime,
		TransitionTime:     transitionTime,
		MinInterval:        minInterval,
		MaxInterval:        maxInterval,
		TimeDividend:       timeDividend,
		MaxSupportedSymbol: maxSupportedSymbol,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	// TODO: adjust the default parameters.
	// - prepare time: 1 day.
	return NewParams("[NOT_SET]", 30, 30, 60, 3600, 3600000, 300)
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
