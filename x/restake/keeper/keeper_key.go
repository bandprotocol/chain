package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetKeysIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyStoreKeyPrefix)
}

func (k Keeper) GetKeys(ctx sdk.Context) (keys []types.Key) {
	iterator := k.GetKeysIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var key types.Key
		k.cdc.MustUnmarshal(iterator.Value(), &key)
		keys = append(keys, key)
	}

	return keys
}

func (k Keeper) HasKey(ctx sdk.Context, keyName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.KeyStoreKey(keyName))
}

func (k Keeper) GetKey(ctx sdk.Context, keyName string) (types.Key, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.KeyStoreKey(keyName))
	if bz == nil {
		return types.Key{}, types.ErrKeyNotFound.Wrapf("failed to get key with name: %s", keyName)
	}

	var f types.Key
	k.cdc.MustUnmarshal(bz, &f)

	return f, nil
}

func (k Keeper) SetKey(ctx sdk.Context, key types.Key) {
	ctx.KVStore(k.storeKey).Set(types.KeyStoreKey(key.Name), k.cdc.MustMarshal(&key))
}

func (k Keeper) DeleteKey(ctx sdk.Context, keyName string) {
	ctx.KVStore(k.storeKey).Delete(types.KeyStoreKey(keyName))
}
