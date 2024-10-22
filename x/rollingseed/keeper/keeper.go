package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/rollingseed/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
}

func NewKeeper(
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		storeKey: storeKey,
	}
}

// SetRollingSeed sets the rolling seed value to be provided value.
func (k Keeper) SetRollingSeed(ctx sdk.Context, rollingSeed []byte) {
	ctx.KVStore(k.storeKey).Set(types.RollingSeedStoreKey, rollingSeed)
}

// GetRollingSeed returns the current rolling seed value.
func (k Keeper) GetRollingSeed(ctx sdk.Context) []byte {
	return ctx.KVStore(k.storeKey).Get(types.RollingSeedStoreKey)
}
