package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// GetParams returns the current x/tss module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	bz := ctx.KVStore(k.storeKey).Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the x/tss module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.ParamsKey, k.cdc.MustMarshal(&p))

	return nil
}
