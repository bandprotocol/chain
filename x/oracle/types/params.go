package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// nolint
const (
	// Each value below is the default value for each parameter when generating the default
	// genesis file. See comments in types.proto for explanation for each parameter.
	DefaultMaxRawRequestCount      = uint64(16)
	DefaultMaxAskCount             = uint64(16)
	DefaultMaxCalldataSize         = uint64(256) // 256B
	DefaultMaxReportDataSize       = uint64(512) // 512B
	DefaultExpirationBlockCount    = uint64(100)
	DefaultBaseRequestGas          = uint64(50000)
	DefaultPerValidatorRequestGas  = uint64(0)
	DefaultSamplingTryCount        = uint64(3)
	DefaultOracleRewardPercentage  = uint64(40)
	DefaultInactivePenaltyDuration = uint64(10 * time.Minute)
	DefaultIBCRequestEnabled       = true
)

// NewParams creates a new parameter configuration for the oracle module
func NewParams(
	maxRawRequestCount, maxAskCount, maxCalldataSize, maxReportDataSize, expirationBlockCount, baseRequestGas, perValidatorRequestGas,
	samplingTryCount, oracleRewardPercentage, inactivePenaltyDuration uint64,
	ibcRequestEnabled bool,
) Params {
	return Params{
		MaxRawRequestCount:      maxRawRequestCount,
		MaxAskCount:             maxAskCount,
		MaxCalldataSize:         maxCalldataSize,
		MaxReportDataSize:       maxReportDataSize,
		ExpirationBlockCount:    expirationBlockCount,
		BaseOwasmGas:            baseRequestGas,
		PerValidatorRequestGas:  perValidatorRequestGas,
		SamplingTryCount:        samplingTryCount,
		OracleRewardPercentage:  oracleRewardPercentage,
		InactivePenaltyDuration: inactivePenaltyDuration,
		IBCRequestEnabled:       ibcRequestEnabled,
	}
}

// DefaultParams defines the default parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxRawRequestCount,
		DefaultMaxAskCount,
		DefaultMaxCalldataSize,
		DefaultMaxReportDataSize,
		DefaultExpirationBlockCount,
		DefaultBaseRequestGas,
		DefaultPerValidatorRequestGas,
		DefaultSamplingTryCount,
		DefaultOracleRewardPercentage,
		DefaultInactivePenaltyDuration,
		DefaultIBCRequestEnabled,
	)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if err := validateUint64("max raw request count", true)(p.MaxRawRequestCount); err != nil {
		return err
	}
	if err := validateUint64("max ask count", true)(p.MaxAskCount); err != nil {
		return err
	}
	if err := validateUint64("max calldata size", true)(p.MaxCalldataSize); err != nil {
		return err
	}
	if err := validateUint64("max report data size", true)(p.MaxReportDataSize); err != nil {
		return err
	}
	if err := validateUint64("expiration block count", true)(p.ExpirationBlockCount); err != nil {
		return err
	}
	if err := validateUint64("base request gas", false)(p.BaseOwasmGas); err != nil {
		return err
	}
	if err := validateUint64("per validator request gas", false)(p.PerValidatorRequestGas); err != nil {
		return err
	}
	if err := validateUint64("sampling try count", true)(p.SamplingTryCount); err != nil {
		return err
	}
	if err := validateUint64("oracle reward percentage", false)(p.OracleRewardPercentage); err != nil {
		return err
	}
	if err := validateUint64("inactive penalty duration", false)(p.InactivePenaltyDuration); err != nil {
		return err
	}
	if err := validateBool()(p.IBCRequestEnabled); err != nil {
		return err
	}

	return nil
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateUint64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}

func validateBool() func(interface{}) error {
	return func(i interface{}) error {
		_, ok := i.(bool)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		return nil
	}
}
