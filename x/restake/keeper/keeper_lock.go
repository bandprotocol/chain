package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetLocksIterator(ctx sdk.Context, address sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.LockStoreKeyPrefix)
}

func (k Keeper) GetLocks(ctx sdk.Context, address sdk.AccAddress) (locks []types.Lock) {
	iterator := k.GetKeysIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lock types.Lock
		k.cdc.MustUnmarshal(iterator.Value(), &lock)
		locks = append(locks, lock)
	}

	return locks
}

func (k Keeper) HasLock(ctx sdk.Context, address sdk.AccAddress, keyName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.LockStoreKey(address, keyName))
}

func (k Keeper) GetLock(ctx sdk.Context, address sdk.AccAddress, keyName string) (types.Lock, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.LockStoreKey(address, keyName))
	if bz == nil {
		return types.Lock{}, types.ErrLockNotFound.Wrapf(
			"failed to get lock of %s with key name: %s",
			address.String(),
			keyName,
		)
	}

	var f types.Lock
	k.cdc.MustUnmarshal(bz, &f)

	return f, nil
}

func (k Keeper) SetLock(ctx sdk.Context, lock types.Lock) {
	address := sdk.MustAccAddressFromBech32(lock.Address)
	ctx.KVStore(k.storeKey).Set(types.LockStoreKey(address, lock.Key), k.cdc.MustMarshal(&lock))
}

func (k Keeper) DeleteLock(ctx sdk.Context, address sdk.AccAddress, keyName string) {
	ctx.KVStore(k.storeKey).Delete(types.LockStoreKey(address, keyName))
}
