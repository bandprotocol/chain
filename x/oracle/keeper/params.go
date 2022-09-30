package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// SetParams sets oracle parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetParams returns all current parameters as a types.Params instance.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.NewParams(
		k.MaxRawRequestCount(ctx),
		k.MaxAskCount(ctx),
		k.MaxCalldataSize(ctx),
		k.MaxReportDataSize(ctx),
		k.ExpirationBlockCount(ctx),
		k.BaseOwasmGas(ctx),
		k.PerValidatorRequestGas(ctx),
		k.SamplingTryCount(ctx),
		k.OracleRewardPercentage(ctx),
		k.InactivePenaltyDuration(ctx),
		k.IBCRequestEnabled(ctx),
	)
}

// MaxRawRequestCount - Maximum number of raw request allowed
func (k Keeper) MaxRawRequestCount(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxRawRequestCount, &res)
	return
}

// MaxAskCount - Maximum number of validators allowed to fulfill the request
func (k Keeper) MaxAskCount(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxAskCount, &res)
	return
}

// MaxCalldataSize - Maximum size limit of calldata (bytes) in a request
func (k Keeper) MaxCalldataSize(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxCalldataSize, &res)
	return
}

// MaxReportDataSize - Maximum size limit of report data (bytes) in a report
func (k Keeper) MaxReportDataSize(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxReportDataSize, &res)
	return
}

// ExpirationBlockCount - number of blocks allowed to fulfill oracle requests
// If current block height is ahead of the block height of an oracle request
// for more than this param value, then the request is considered expired
func (k Keeper) ExpirationBlockCount(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyExpirationBlockCount, &res)
	return
}

// BaseOwasmGas - amount of gas consumed by owasm required as a baseline
func (k Keeper) BaseOwasmGas(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyBaseOwasmGas, &res)
	return
}

// PerValidatorRequestGas - amount of gas consumed when preparing request
// It is the amount per validator where total gas depends on ask count
func (k Keeper) PerValidatorRequestGas(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyPerValidatorRequestGas, &res)
	return
}

// SamplingTryCount - number of tries when sampling validator set
func (k Keeper) SamplingTryCount(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeySamplingTryCount, &res)
	return
}

// OracleRewardPercentage - reward ratio used when allocating reward
// for active validators
func (k Keeper) OracleRewardPercentage(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyOracleRewardPercentage, &res)
	return
}

// InactivePenaltyDuration - time duration that need to wait when
// the validator marked as inactive oracle provider, before
// re-activate as oracle provider
func (k Keeper) InactivePenaltyDuration(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyInactivePenaltyDuration, &res)
	return
}

// IBCRequestEnabled - a flag indicating whether sending oracle request via
// IBC is allowed or not
func (k Keeper) IBCRequestEnabled(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyIBCRequestEnabled, &res)
	return
}
