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
	DefaultMaxRawRequestCount      = uint64(12)
	DefaultMaxAskCount             = uint64(16)
	DefaultExpirationBlockCount    = uint64(100)
	DefaultBaseRequestGas          = uint64(150000)
	DefaultPerValidatorRequestGas  = uint64(30000)
	DefaultSamplingTryCount        = uint64(3)
	DefaultOracleRewardPercentage  = uint64(70)
	DefaultInactivePenaltyDuration = uint64(10 * time.Minute)
	DefaultMaxDataSize             = uint64(1 * 1024) // 1 KB
	DefaultMaxCalldataSize         = uint64(1 * 1024) // 1 KB
	DefaultPrepareGas              = uint64(40000)
	DefaultExecuteGas              = uint64(300000)
	DefaultRewardThresholdBlocks   = uint64(28820)
	DefaultDataProviderRewardDenom = "minigeo"
	DefaultDataRequesterFeeDenom   = "loki"
)

var (
	DefaultDataProviderRewardPerByte = sdk.NewCoins(sdk.NewInt64Coin(DefaultDataProviderRewardDenom, 1000000)) // 1 * 10^6
	DefaultDataRequesterFeeDenoms    = []string{DefaultDataRequesterFeeDenom}
	DefaultFeeLimit                  = sdk.NewCoins()
	DefaultRewardThresholdAmount     = sdk.NewCoins(sdk.NewInt64Coin(DefaultDataProviderRewardDenom, 200000000000)) // 200000 * 10^6
	DefaultRewardDecreasingFraction  = sdk.NewDec(1).Quo(sdk.NewDec(20))
)

// nolint
var (
	// Each value below is the key to store the respective oracle module parameter. See comments
	// in types.proto for explanation for each parameter.
	KeyMaxRawRequestCount          = []byte("MaxRawRequestCount")
	KeyMaxAskCount                 = []byte("MaxAskCount")
	KeyExpirationBlockCount        = []byte("ExpirationBlockCount")
	KeyBaseOwasmGas                = []byte("BaseOwasmGas")
	KeyPerValidatorRequestGas      = []byte("PerValidatorRequestGas")
	KeySamplingTryCount            = []byte("SamplingTryCount")
	KeyOracleRewardPercentage      = []byte("OracleRewardPercentage")
	KeyInactivePenaltyDuration     = []byte("InactivePenaltyDuration")
	KeyMaxDataSize                 = []byte("MaxDataSize")
	KeyMaxCalldataSize             = []byte("MaxCalldataSize")
	KeyDataProviderRewardPerByte   = []byte("DataProviderRewardPerByte")
	KeyRewardDecreasingFraction    = []byte("RewardDecreasingFraction")
	KeyDataProviderRewardThreshold = []byte("DataProviderRewardThreshold")
	KeyDataRequesterFeeDenoms      = []byte("DataRequesterFeeDenoms")
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
	dataProviderRewardPerByte sdk.Coins, dataProviderRewardThreshold RewardThreshold, rewardDecreasingFraction sdk.Dec,
	dataRequesterFeeDenoms []string,
) Params {
	return Params{
		MaxRawRequestCount:          maxRawRequestCount,
		MaxAskCount:                 maxAskCount,
		ExpirationBlockCount:        expirationBlockCount,
		BaseOwasmGas:                baseRequestGas,
		PerValidatorRequestGas:      perValidatorRequestGas,
		SamplingTryCount:            samplingTryCount,
		OracleRewardPercentage:      oracleRewardPercentage,
		InactivePenaltyDuration:     inactivePenaltyDuration,
		MaxDataSize:                 maxDataSize,
		MaxCalldataSize:             maxCallDataSize,
		DataProviderRewardPerByte:   dataProviderRewardPerByte,
		DataProviderRewardThreshold: dataProviderRewardThreshold,
		RewardDecreasingFraction:    rewardDecreasingFraction,
		DataRequesterFeeDenoms:      dataRequesterFeeDenoms,
	}
}

func NewRewardThreshold(amount sdk.Coins, blocks uint64) RewardThreshold {
	return RewardThreshold{
		Amount: amount,
		Blocks: blocks,
	}
}

func DefaultRewardThreshold() RewardThreshold {
	return RewardThreshold{
		Amount: DefaultRewardThresholdAmount,
		Blocks: DefaultRewardThresholdBlocks,
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
		paramtypes.NewParamSetPair(KeyDataProviderRewardPerByte, &p.DataProviderRewardPerByte, validateDataProviderReward),
		paramtypes.NewParamSetPair(KeyRewardDecreasingFraction, &p.RewardDecreasingFraction, validateRewardDecreasingFraction),
		paramtypes.NewParamSetPair(KeyDataProviderRewardThreshold, &p.DataProviderRewardThreshold, validateRewardThreshold),
		paramtypes.NewParamSetPair(KeyDataRequesterFeeDenoms, &p.DataRequesterFeeDenoms, validateDataRequesterFeeDenoms),
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
		DefaultRewardThreshold(),
		DefaultRewardDecreasingFraction,
		DefaultDataRequesterFeeDenoms,
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

func validateDataProviderReward(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsAnyNegative() {
		return fmt.Errorf("data provider reward must be positive: %v", v)
	}

	return nil
}

func validateRewardDecreasingFraction(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("reward decresing fraction must be positive: %v", v)
	}
	if v.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reward decresing fraction must be less or equal to 1 %v", v)
	}
	return nil
}

func validateRewardThreshold(i interface{}) error {
	v, ok := i.(RewardThreshold)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.Amount.IsAnyNegative() {
		return fmt.Errorf("reward threshold amount must be positive: %v", v.Amount)
	}
	if v.Amount.IsZero() {
		return fmt.Errorf("reward threshold amount must be greater than zero: %v", v.Amount)
	}
	if v.Blocks <= 0 {
		return fmt.Errorf("reward threshold blocks count must be greater than zero: %v", v.Blocks)
	}
	return nil
}

func validateDataRequesterFeeDenoms(i interface{}) error {
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, d := range v {
		if err := sdk.ValidateDenom(d); err != nil {
			return fmt.Errorf("denoms must be valid: %s, error: %w", d, err)
		}
	}
	return nil
}
