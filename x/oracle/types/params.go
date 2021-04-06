package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"

	"gopkg.in/yaml.v2"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// nolint
const (
	// Each value below is the default value for each parameter when generating the default
	// genesis file. See comments in types.proto for explanation for each parameter.
	DefaultMaxRawRequestCount                = uint64(12)
	DefaultMaxAskCount                       = uint64(16)
	DefaultExpirationBlockCount              = uint64(100)
	DefaultBaseRequestGas                    = uint64(150000)
	DefaultPerValidatorRequestGas            = uint64(30000)
	DefaultSamplingTryCount                  = uint64(3)
	DefaultOracleRewardPercentage            = uint64(70)
	DefaultInactivePenaltyDuration           = uint64(10 * time.Minute)
	DefaultMaxDataSize                       = uint64(1 * 1024) // 1 KB
	DefaultMaxCalldataSize                   = uint64(1 * 1024) // 1 KB
	DefaultDataProviderRewardDenom           = "geo"
	DefaultDataRequesterBasicFeeDenom        = "odin"
	DefaultPrepareGas                 uint64 = 40000
	DefaultExecuteGas                 uint64 = 300000
)

var (
	DefaultDataProviderRewardPerByte = sdk.NewInt64DecCoin(DefaultDataProviderRewardDenom, 0)
	DefaultDataRequesterBasicFee     = sdk.NewInt64Coin(DefaultDataRequesterBasicFeeDenom, 0)
	DefaultFeeLimit                  = sdk.NewCoins()
)

// nolint
var (
	// Each value below is the key to store the respective oracle module parameter. See comments
	// in types.proto for explanation for each parameter.
	KeyMaxRawRequestCount        = []byte("MaxRawRequestCount")
	KeyMaxAskCount               = []byte("MaxAskCount")
	KeyExpirationBlockCount      = []byte("ExpirationBlockCount")
	KeyBaseOwasmGas              = []byte("BaseOwasmGas")
	KeyPerValidatorRequestGas    = []byte("PerValidatorRequestGas")
	KeySamplingTryCount          = []byte("SamplingTryCount")
	KeyOracleRewardPercentage    = []byte("OracleRewardPercentage")
	KeyInactivePenaltyDuration   = []byte("InactivePenaltyDuration")
	KeyMaxDataSize               = []byte("MaxDataSize")
	KeyMaxCalldataSize           = []byte("MaxCalldataSize")
	KeyDataProviderRewardPerByte = []byte("DataProviderRewardPerByte")
	KeyDataRequesterBasicFee     = []byte("DataRequesterBasicFee")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable for oracle module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the oracle module
func NewParams(
	maxRawRequestCount, maxAskCount, expirationBlockCount, baseRequestGas, perValidatorRequestGas,
	samplingTryCount, oracleRewardPercentage, inactivePenaltyDuration, maxDataSize, maxCallDataSize uint64,
	dataProviderRewardPerByte sdk.DecCoin, dataRequesterBasicFee sdk.Coin,
) Params {
	return Params{
		MaxRawRequestCount:        maxRawRequestCount,
		MaxAskCount:               maxAskCount,
		ExpirationBlockCount:      expirationBlockCount,
		BaseOwasmGas:              baseRequestGas,
		PerValidatorRequestGas:    perValidatorRequestGas,
		SamplingTryCount:          samplingTryCount,
		OracleRewardPercentage:    oracleRewardPercentage,
		InactivePenaltyDuration:   inactivePenaltyDuration,
		MaxDataSize:               maxDataSize,
		MaxCalldataSize:           maxCallDataSize,
		DataProviderRewardPerByte: dataProviderRewardPerByte,
		DataRequesterBasicFee:     dataRequesterBasicFee,
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
		paramtypes.NewParamSetPair(KeyMaxDataSize, &p.MaxDataSize, validateUint64("max data size", true)),
		paramtypes.NewParamSetPair(KeyMaxCalldataSize, &p.MaxCalldataSize, validateUint64("max calldata size", true)),
		paramtypes.NewParamSetPair(KeyDataProviderRewardPerByte, &p.DataProviderRewardPerByte, validateDataProviderRewardPerByte),
		paramtypes.NewParamSetPair(KeyDataRequesterBasicFee, &p.DataRequesterBasicFee, validateDataRequesterFee),
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
		DefaultMaxDataSize,
		DefaultMaxCalldataSize,
		DefaultDataProviderRewardPerByte,
		DefaultDataRequesterBasicFee,
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

func validateDataProviderRewardPerByte(i interface{}) error {
	v, ok := i.(sdk.DecCoin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Amount.IsNegative() {
		return fmt.Errorf("data provider reward must be positive: %v", v)
	}
	return nil
}

func validateDataRequesterFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Amount.IsNegative() {
		return fmt.Errorf("data requester fee must be positive: %v", v)
	}
	return nil
}
