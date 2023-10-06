/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// nolint
// Parameter store keys
var (
	// Each value below is the key to store the respective oracle module parameter. See comments
	// in types.proto for explanation for each parameter.
	KeyMaxRawRequestCount      = []byte("MaxRawRequestCount")
	KeyMaxAskCount             = []byte("MaxAskCount")
	KeyMaxCalldataSize         = []byte("MaxCalldataSize")
	KeyMaxReportDataSize       = []byte("MaxReportDataSize")
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

// ParamSetPairs implements the paramtypes.ParamSet interface for Params.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyMaxRawRequestCount,
			&p.MaxRawRequestCount,
			validateUint64("max data source count", true),
		),
		paramtypes.NewParamSetPair(KeyMaxAskCount, &p.MaxAskCount, validateUint64("max ask count", true)),
		paramtypes.NewParamSetPair(KeyMaxCalldataSize, &p.MaxCalldataSize, validateUint64("max calldata size", true)),
		paramtypes.NewParamSetPair(
			KeyMaxReportDataSize,
			&p.MaxReportDataSize,
			validateUint64("max report data size", true),
		),
		paramtypes.NewParamSetPair(
			KeyExpirationBlockCount,
			&p.ExpirationBlockCount,
			validateUint64("expiration block count", true),
		),
		paramtypes.NewParamSetPair(KeyBaseOwasmGas, &p.BaseOwasmGas, validateUint64("base request gas", false)),
		paramtypes.NewParamSetPair(
			KeyPerValidatorRequestGas,
			&p.PerValidatorRequestGas,
			validateUint64("per validator request gas", false),
		),
		paramtypes.NewParamSetPair(
			KeySamplingTryCount,
			&p.SamplingTryCount,
			validateUint64("sampling try count", true),
		),
		paramtypes.NewParamSetPair(
			KeyOracleRewardPercentage,
			&p.OracleRewardPercentage,
			validateUint64("oracle reward percentage", false),
		),
		paramtypes.NewParamSetPair(
			KeyInactivePenaltyDuration,
			&p.InactivePenaltyDuration,
			validateUint64("inactive penalty duration", false),
		),
		paramtypes.NewParamSetPair(KeyIBCRequestEnabled, &p.IBCRequestEnabled, validateBool()),
	}
}
