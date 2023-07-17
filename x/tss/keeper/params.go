package keeper

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params from the global param store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxGroupSize(ctx),
		k.MaxDESize(ctx),
		k.CreatingPeriod(ctx),
		k.SigningPeriod(ctx),
	)
}

// SetParams set the params to the global param store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// MaxGroupSize returns the current MaxGroupSize from the global param store
func (k Keeper) MaxGroupSize(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMaxGroupSize, &res)
	return
}

// MaxDESize returns the current MaxDESize from the global param store
func (k Keeper) MaxDESize(ctx sdk.Context) (res uint64) {
	k.paramSpace.Get(ctx, types.KeyMaxDESize, &res)
	return
}

// CreatingPeriod returns the current CreatingPeriod from the global param store
func (k Keeper) CreatingPeriod(ctx sdk.Context) (res int64) {
	k.paramSpace.Get(ctx, types.KeyCreatingPeriod, &res)
	return
}

// SigningPeriod returns the current SigningPeriod from the global param store
func (k Keeper) SigningPeriod(ctx sdk.Context) (res int64) {
	k.paramSpace.Get(ctx, types.KeySigningPeriod, &res)
	return
}
