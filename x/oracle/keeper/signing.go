package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// SetSigningID sets the key-value pair of the request ID to signing ID in the store.
func (k Keeper) SetSigningResult(ctx sdk.Context, rid types.RequestID, signingResult types.SigningResult) {
	ctx.KVStore(k.storeKey).Set(types.SigningResultStoreKey(rid), k.cdc.MustMarshal(&signingResult))
}

// GetSigningResult retrieves the signing result associated with the given request ID from the store.
func (k Keeper) GetSigningResult(ctx sdk.Context, rid types.RequestID) (types.SigningResult, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningResultStoreKey(rid))

	// Check if the value is not found in the store
	if bz == nil {
		return types.SigningResult{}, types.ErrSigningResultNotFound.Wrapf("id: %d", rid)
	}

	var result types.SigningResult
	k.cdc.MustUnmarshal(bz, &result)
	return result, nil
}
