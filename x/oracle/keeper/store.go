package keeper

import (
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// get the oracle pool info
func (k Keeper) GetOraclePool(ctx sdk.Context) (oraclePool types.OraclePool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.OraclePoolStoreKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &oraclePool)
	return
}

// set the oracle pool info
func (k Keeper) SetOraclePool(ctx sdk.Context, oraclePool types.OraclePool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(&oraclePool)
	store.Set(types.OraclePoolStoreKey, b)
}
