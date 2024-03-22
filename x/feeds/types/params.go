package types

import (
	"gopkg.in/yaml.v2"
)

const (
	// Default values for Params
	DefaultAllowDiffTime     = int64(30)
	DefaultTransitionTime    = int64(30)
	DefaultMinInterval       = int64(60)
	DefaultMaxInterval       = int64(3600)
	DefaultPowerThreshold    = int64(1_000_000_000)
	DefaultMaxSupportedFeeds = int64(100)
)

// NewParams creates a new Params instance
func NewParams(
	admin string,
	allowDiffTime int64,
	transitionTime int64,
	minInterval int64,
	maxInterval int64,
	powerThreshold int64,
	maxSupportedFeeds int64,
) Params {
	return Params{
		Admin:             admin,
		AllowDiffTime:     allowDiffTime,
		TransitionTime:    transitionTime,
		MinInterval:       minInterval,
		MaxInterval:       maxInterval,
		PowerThreshold:    powerThreshold,
		MaxSupportedFeeds: maxSupportedFeeds,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		"[NOT_SET]",
		DefaultAllowDiffTime,
		DefaultTransitionTime,
		DefaultMinInterval,
		DefaultMaxInterval,
		DefaultPowerThreshold,
		DefaultMaxSupportedFeeds,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateString("admin", true)(p.Admin); err != nil {
		return err
	}
	if err := validateInt64("allow diff time", true)(p.AllowDiffTime); err != nil {
		return err
	}
	if err := validateInt64("transition time", true)(p.TransitionTime); err != nil {
		return err
	}
	if err := validateInt64("min interval", true)(p.MinInterval); err != nil {
		return err
	}
	if err := validateInt64("max interval", true)(p.MaxInterval); err != nil {
		return err
	}
	if err := validateInt64("power threshold", true)(p.PowerThreshold); err != nil {
		return err
	}
	if err := validateInt64("max supported feeds", true)(p.MaxSupportedFeeds); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
