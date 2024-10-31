package types

import (
	"fmt"

	"cosmossdk.io/math"
)

var (
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
	DefaultPriceQuorum                = math.LegacyNewDecWithPrec(30, 2)
	DefaultMaxSignalIDsPerSigning     = uint64(10)
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
	priceQuorum string,
	maxSignalIDsPerSigning uint64,
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
		PriceQuorum:                   priceQuorum,
		MaxSignalIDsPerSigning:        maxSignalIDsPerSigning,
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
		DefaultPriceQuorum.String(),
		DefaultMaxSignalIDsPerSigning,
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

	priceQuorum, err := math.LegacyNewDecFromStr(p.PriceQuorum)
	if err != nil {
		return fmt.Errorf("invalid price quorum string: %w", err)
	}
	if priceQuorum.IsNegative() {
		return fmt.Errorf("price quorom cannot be negative: %s", p.PriceQuorum)
	}
	if priceQuorum.GT(math.LegacyOneDec()) {
		return fmt.Errorf("price quorom too large: %s", p.PriceQuorum)
	}

	if err := validateUint64("max signal IDs per signing", false, p.MaxSignalIDsPerSigning); err != nil {
		return err
	}

	return nil
}
