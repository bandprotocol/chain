package types

import (
	"gopkg.in/yaml.v2"
)

const (
	// Default values for Params
	DefaultAllowableBlockTimeDiscrepancy = int64(60)
	DefaultGracePeriod                   = int64(30)
	DefaultMinInterval                   = int64(60)
	DefaultMaxInterval                   = int64(3600)
	DefaultPowerStepThreshold            = int64(1_000_000_000)
	DefaultMaxSupportedFeeds             = int64(300)
	DefaultCooldownTime                  = int64(30)
	DefaultMinDeviationBasisPoint        = int64(50)
	DefaultMaxDeviationBasisPoint        = int64(3000)
	// estimated from block time of 3 seconds, aims for 1 day update
	DefaultSupportedFeedsUpdateInterval = uint64(28800)
)

// NewParams creates a new Params instance
func NewParams(
	admin string,
	allowableBlockTimeDiscrepancy int64,
	gracePeriod int64,
	minInterval int64,
	maxInterval int64,
	powerStepThreshold int64,
	maxSupportedFeeds int64,
	cooldownTime int64,
	minDeviationBasisPoint int64,
	maxDeviationBasisPoint int64,
	supportedFeedsUpdateInterval uint64,
) Params {
	return Params{
		Admin:                         admin,
		AllowableBlockTimeDiscrepancy: allowableBlockTimeDiscrepancy,
		GracePeriod:                   gracePeriod,
		MinInterval:                   minInterval,
		MaxInterval:                   maxInterval,
		PowerStepThreshold:            powerStepThreshold,
		MaxSupportedFeeds:             maxSupportedFeeds,
		CooldownTime:                  cooldownTime,
		MinDeviationBasisPoint:        minDeviationBasisPoint,
		MaxDeviationBasisPoint:        maxDeviationBasisPoint,
		SupportedFeedsUpdateInterval:  supportedFeedsUpdateInterval,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		"[NOT_SET]",
		DefaultAllowableBlockTimeDiscrepancy,
		DefaultGracePeriod,
		DefaultMinInterval,
		DefaultMaxInterval,
		DefaultPowerStepThreshold,
		DefaultMaxSupportedFeeds,
		DefaultCooldownTime,
		DefaultMinDeviationBasisPoint,
		DefaultMaxDeviationBasisPoint,
		DefaultSupportedFeedsUpdateInterval,
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
	if err := validateInt64("grace period", true, p.GracePeriod); err != nil {
		return err
	}
	if err := validateInt64("min interval", true, p.MinInterval); err != nil {
		return err
	}
	if err := validateInt64("max interval", true, p.MaxInterval); err != nil {
		return err
	}
	if err := validateInt64("power threshold", true, p.PowerStepThreshold); err != nil {
		return err
	}
	if err := validateInt64("max supported feeds", true, p.MaxSupportedFeeds); err != nil {
		return err
	}
	if err := validateInt64("cooldown time", true, p.CooldownTime); err != nil {
		return err
	}
	if err := validateInt64("min deviation basis point", true, p.MinDeviationBasisPoint); err != nil {
		return err
	}
	if err := validateInt64("max deviation basis point", true, p.MaxDeviationBasisPoint); err != nil {
		return err
	}
	if err := validateUint64("supported feeds update interval", true, p.SupportedFeedsUpdateInterval); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
