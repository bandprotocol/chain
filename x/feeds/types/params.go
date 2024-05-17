package types

import (
	"gopkg.in/yaml.v2"
)

const (
	// Default values for Params
	DefaultAllowableBlockTimeDiscrepancy = int64(60)
	DefaultTransitionTime                = int64(30)
	DefaultMinInterval                   = int64(60)
	DefaultMaxInterval                   = int64(3600)
	DefaultPowerThreshold                = int64(1_000_000_000)
	DefaultMaxSupportedFeeds             = int64(300)
	DefaultCooldownTime                  = int64(30)
	DefaultMinDeviationInThousandth      = int64(5)
	DefaultMaxDeviationInThousandth      = int64(300)
	DefaultMaxSignalIDCharacters         = uint64(256)
	DefaultBlocksPerFeedsUpdate          = uint64(
		28800,
	) // estimated from block time of 3 seconds, aims for 1 day update
)

// NewParams creates a new Params instance
func NewParams(
	admin string,
	allowableBlockTimeDiscrepancy int64,
	transitionTime int64,
	minInterval int64,
	maxInterval int64,
	powerThreshold int64,
	maxSupportedFeeds int64,
	cooldownTime int64,
	minDeviationInThousandth int64,
	maxDeviationInThousandth int64,
	maxSignalIDCharacters uint64,
	blocksPerFeedsUpdate uint64,
) Params {
	return Params{
		Admin:                         admin,
		AllowableBlockTimeDiscrepancy: allowableBlockTimeDiscrepancy,
		TransitionTime:                transitionTime,
		MinInterval:                   minInterval,
		MaxInterval:                   maxInterval,
		PowerThreshold:                powerThreshold,
		MaxSupportedFeeds:             maxSupportedFeeds,
		CooldownTime:                  cooldownTime,
		MinDeviationInThousandth:      minDeviationInThousandth,
		MaxDeviationInThousandth:      maxDeviationInThousandth,
		MaxSignalIDCharacters:         maxSignalIDCharacters,
		BlocksPerFeedsUpdate:          blocksPerFeedsUpdate,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		"[NOT_SET]",
		DefaultAllowableBlockTimeDiscrepancy,
		DefaultTransitionTime,
		DefaultMinInterval,
		DefaultMaxInterval,
		DefaultPowerThreshold,
		DefaultMaxSupportedFeeds,
		DefaultCooldownTime,
		DefaultMinDeviationInThousandth,
		DefaultMaxDeviationInThousandth,
		DefaultMaxSignalIDCharacters,
		DefaultBlocksPerFeedsUpdate,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateString("admin", true, p.Admin); err != nil {
		return err
	}
	if err := validateInt64("allowable block time discrepancy", true, p.AllowableBlockTimeDiscrepancy); err != nil {
		return err
	}
	if err := validateInt64("transition time", true, p.TransitionTime); err != nil {
		return err
	}
	if err := validateInt64("min interval", true, p.MinInterval); err != nil {
		return err
	}
	if err := validateInt64("max interval", true, p.MaxInterval); err != nil {
		return err
	}
	if err := validateInt64("power threshold", true, p.PowerThreshold); err != nil {
		return err
	}
	if err := validateInt64("max supported feeds", true, p.MaxSupportedFeeds); err != nil {
		return err
	}
	if err := validateInt64("cooldown time", true, p.CooldownTime); err != nil {
		return err
	}
	if err := validateInt64("min deviation in thousandth", true, p.MinDeviationInThousandth); err != nil {
		return err
	}
	if err := validateInt64("max deviation in thousandth", true, p.MaxDeviationInThousandth); err != nil {
		return err
	}
	if err := validateUint64("max signal id characters", true, p.MaxSignalIDCharacters); err != nil {
		return err
	}
	if err := validateUint64("blocks per feeds update", true, p.BlocksPerFeedsUpdate); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
