package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// nolint
const (
	// Each value below is the default value for each parameter when generating the default
	// genesis file. See comments in types.proto for explanation for each parameter.
	DefaultMaxRawRequestCount      = uint64(12)
	DefaultMaxAskCount             = uint64(16)
	DefaultExpirationBlockCount    = uint64(100)
	DefaultBaseRequestGas          = uint64(20000)
	DefaultPerValidatorRequestGas  = uint64(30000)
	DefaultSamplingTryCount        = uint64(3)
	DefaultOracleRewardPercentage  = uint64(70)
	DefaultInactivePenaltyDuration = uint64(10 * time.Minute)
	DefaultIBCRequestEnabled       = true
)

// nolint
var (
	// Each value below is the key to store the respective oracle module parameter. See comments
	// in types.proto for explanation for each parameter.
	KeyMaxRawRequestCount      = []byte("MaxRawRequestCount")
	KeyMaxAskCount             = []byte("MaxAskCount")
	KeyExpirationBlockCount    = []byte("ExpirationBlockCount")
	KeyBaseOwasmGas            = []byte("BaseOwasmGas")
	KeyPerValidatorRequestGas  = []byte("PerValidatorRequestGas")
	KeySamplingTryCount        = []byte("SamplingTryCount")
	KeyOracleRewardPercentage  = []byte("OracleRewardPercentage")
	KeyInactivePenaltyDuration = []byte("InactivePenaltyDuration")
	KeyIBCRequestEnabled       = []byte("IBCRequestEnabled")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable for oracle module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the oracle module
func NewParams(
	maxRawRequestCount, maxAskCount, expirationBlockCount, baseRequestGas, perValidatorRequestGas,
	samplingTryCount, oracleRewardPercentage, inactivePenaltyDuration uint64, ibcRequestEnabled bool,
) Params {
	return Params{
		MaxRawRequestCount:      maxRawRequestCount,
		MaxAskCount:             maxAskCount,
		ExpirationBlockCount:    expirationBlockCount,
		BaseOwasmGas:            baseRequestGas,
		PerValidatorRequestGas:  perValidatorRequestGas,
		SamplingTryCount:        samplingTryCount,
		OracleRewardPercentage:  oracleRewardPercentage,
		InactivePenaltyDuration: inactivePenaltyDuration,
		IBCRequestEnabled:       ibcRequestEnabled,
	}
}

// ParamSetPairs implements the paramtypes.ParamSet interface for Params.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxRawRequestCount, &p.MaxRawRequestCount, validateUint64("max data source count", true)),
		paramtypes.NewParamSetPair(KeyMaxAskCount, &p.MaxAskCount, validateUint64("max ask count", true)),
		paramtypes.NewParamSetPair(KeyExpirationBlockCount, &p.ExpirationBlockCount, validateUint64("expiration block count", true)),
		paramtypes.NewParamSetPair(KeyBaseOwasmGas, &p.BaseOwasmGas, validateUint64("base request gas", false)),
		paramtypes.NewParamSetPair(KeyPerValidatorRequestGas, &p.PerValidatorRequestGas, validateUint64("per validator request gas", false)),
		paramtypes.NewParamSetPair(KeySamplingTryCount, &p.SamplingTryCount, validateUint64("sampling try count", true)),
		paramtypes.NewParamSetPair(KeyOracleRewardPercentage, &p.OracleRewardPercentage, validateUint64("oracle reward percentage", false)),
		paramtypes.NewParamSetPair(KeyInactivePenaltyDuration, &p.InactivePenaltyDuration, validateUint64("inactive penalty duration", false)),
		paramtypes.NewParamSetPair(KeyIBCRequestEnabled, &p.IBCRequestEnabled, validateBool()),
	}
}

// DefaultParams defines the default parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxRawRequestCount,
		DefaultMaxAskCount,
		DefaultExpirationBlockCount,
		DefaultBaseRequestGas,
		DefaultPerValidatorRequestGas,
		DefaultSamplingTryCount,
		DefaultOracleRewardPercentage,
		DefaultInactivePenaltyDuration,
		DefaultIBCRequestEnabled,
	)
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
