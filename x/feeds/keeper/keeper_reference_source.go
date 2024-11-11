package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// GetReferenceSourceConfig gets the current reference source config.
func (k Keeper) GetReferenceSourceConfig(ctx sdk.Context) (rs types.ReferenceSourceConfig) {
	bz := ctx.KVStore(k.storeKey).Get(types.ReferenceSourceConfigStoreKey)
	if bz == nil {
		return rs
	}

	k.cdc.MustUnmarshal(bz, &rs)

	return rs
}

// SetReferenceSourceConfig sets new reference source config to the store.
func (k Keeper) SetReferenceSourceConfig(ctx sdk.Context, rs types.ReferenceSourceConfig) error {
	if err := rs.Validate(); err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.ReferenceSourceConfigStoreKey, k.cdc.MustMarshal(&rs))

	return nil
}
