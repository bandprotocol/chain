package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// SetSigningID sets the key-value pair of the request ID to signing ID in the store.
func (k Keeper) SetSigningID(ctx sdk.Context, rid types.RequestID, sid tss.SigningID) {
	ctx.KVStore(k.storeKey).Set(types.SigningIDStoreKey(rid), sdk.Uint64ToBigEndian(uint64(sid)))
}

// GetSigningID retrieves the signing ID associated with the given request ID from the store.
func (k Keeper) GetSigningID(ctx sdk.Context, rid types.RequestID) (tss.SigningID, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningIDStoreKey(rid))

	// Check if the value is not found in the store
	if bz == nil {
		return 0, sdkerrors.Wrapf(types.ErrResultNotFound, "id: %d", rid)
	}

	return tss.SigningID(sdk.BigEndianToUint64(bz)), nil
}
