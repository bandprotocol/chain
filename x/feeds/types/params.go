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
	DefaultMaxCurrentFeeds               = uint64(300)
	DefaultCooldownTime                  = int64(30)
	DefaultMinDeviationBasisPoint        = int64(50)
	DefaultMaxDeviationBasisPoint        = int64(3000)
	// estimated from block time of 3 seconds, aims for 1 day update
	DefaultCurrentFeedsUpdateInterval = int64(28800)
)

// NewParams creates a new Params instance
func NewParams(
	admin string,
	allowableBlockTimeDiscrepancy int64,
	gracePeriod int64,
	minInterval int64,
	maxInterval int64,
	powerStepThreshold int64,
	maxCurrentFeeds uint64,
	cooldownTime int64,
	minDeviationBasisPoint int64,
	maxDeviationBasisPoint int64,
	currentFeedsUpdateInterval int64,
) Params {
	return Params{
		Admin:                         admin,
		AllowableBlockTimeDiscrepancy: allowableBlockTimeDiscrepancy,
		GracePeriod:                   gracePeriod,
		MinInterval:                   minInterval,
		MaxInterval:                   maxInterval,
		PowerStepThreshold:            powerStepThreshold,
		MaxCurrentFeeds:               maxCurrentFeeds,
		CooldownTime:                  cooldownTime,
		MinDeviationBasisPoint:        minDeviationBasisPoint,
		MaxDeviationBasisPoint:        maxDeviationBasisPoint,
		CurrentFeedsUpdateInterval:    currentFeedsUpdateInterval,
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
		DefaultMaxCurrentFeeds,
		DefaultCooldownTime,
		DefaultMinDeviationBasisPoint,
		DefaultMaxDeviationBasisPoint,
		DefaultCurrentFeedsUpdateInterval,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateString("admin", false, p.Admin); err != nil {
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
	if err := validateUint64("max current feeds", false, p.MaxCurrentFeeds); err != nil {
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
	if err := validateInt64("current feeds update interval", true, p.CurrentFeedsUpdateInterval); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
