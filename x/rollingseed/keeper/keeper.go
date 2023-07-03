package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/rollingseed/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
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
