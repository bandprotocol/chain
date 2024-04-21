package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetOrCreateKey(ctx sdk.Context, keyName string) types.Key {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		key = types.Key{
			Name:            keyName,
			IsActive:        true,
			TotalLock:       sdk.NewInt(0),
			RewardPerShares: sdk.NewDecCoins(),
			CurrentRewards:  sdk.NewDecCoins(),
		}

		k.SetKey(ctx, key)
	}

	return key
}

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

	var key types.Key
	k.cdc.MustUnmarshal(bz, &key)

	return key, nil
}

func (k Keeper) SetKey(ctx sdk.Context, key types.Key) {
	ctx.KVStore(k.storeKey).Set(types.KeyStoreKey(key.Name), k.cdc.MustMarshal(&key))
}

func (k Keeper) DeleteKey(ctx sdk.Context, keyName string) {
	ctx.KVStore(k.storeKey).Delete(types.KeyStoreKey(keyName))
}

func (k Keeper) DeactivateKey(ctx sdk.Context, keyName string) error {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	key.IsActive = false
	k.SetKey(ctx, key)

	return nil
}

func (k Keeper) updateRewardPerShares(ctx sdk.Context, key types.Key) types.Key {
	if key.TotalLock.IsZero() {
		k.addRemainderAmount(ctx, key.CurrentRewards)
	} else {
		key.RewardPerShares = key.RewardPerShares.Add(
			key.CurrentRewards.QuoDecTruncate(sdk.NewDecFromInt(key.TotalLock))...)
		key.CurrentRewards = sdk.NewDecCoins()
		k.SetKey(ctx, key)
	}

	return key
}
