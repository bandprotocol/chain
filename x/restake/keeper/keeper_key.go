package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetOrCreateKey(ctx sdk.Context, keyName string) (types.Key, error) {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		keyAccAddr, err := k.CreateKeyAccount(ctx, keyName)
		if err != nil {
			return types.Key{}, err
		}

		key = types.Key{
			Name:            keyName,
			Address:         keyAccAddr.String(),
			IsActive:        true,
			TotalLock:       sdk.NewInt(0),
			RewardPerShares: sdk.NewDecCoins(),
			CurrentRewards:  sdk.NewDecCoins(),
		}

		k.SetKey(ctx, key)
	}

	return key, nil
}

func (k Keeper) CreateKeyAccount(ctx sdk.Context, key string) (sdk.AccAddress, error) {
	header := ctx.BlockHeader()

	buf := []byte(key)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	moduleCred, err := authtypes.NewModuleCredential(types.ModuleName, []byte(types.KeyAccountsKey), buf)
	if err != nil {
		return nil, err
	}

	keyAccAddr := sdk.AccAddress(moduleCred.Address())

	// This should not happen
	if acc := k.authKeeper.GetAccount(ctx, keyAccAddr); acc != nil {
		return nil, types.ErrAccountAlreadyExist.Wrapf(
			"existing account for newly generated key account address %s",
			keyAccAddr.String(),
		)
	}

	keyAcc, err := authtypes.NewBaseAccountWithPubKey(moduleCred)
	if err != nil {
		return nil, err
	}

	k.authKeeper.NewAccount(ctx, keyAcc)
	k.authKeeper.SetAccount(ctx, keyAcc)

	return keyAccAddr, nil
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

func (k Keeper) MustGetKey(ctx sdk.Context, keyName string) types.Key {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		panic(err)
	}

	return key
}

func (k Keeper) IsActiveKey(ctx sdk.Context, keyName string) bool {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return false
	}

	return key.IsActive
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

func (k Keeper) DeactivateKey(ctx sdk.Context, keyName string) error {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	if !key.IsActive {
		return types.ErrKeyAlreadyDeactivated
	}

	key.IsActive = false
	k.SetKey(ctx, key)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeactivateKey,
			sdk.NewAttribute(types.AttributeKeyKey, keyName),
		),
	)

	return nil
}

func (k Keeper) ProcessKey(ctx sdk.Context, key types.Key) types.Key {
	if key.TotalLock.IsZero() {
		return key
	}

	key.RewardPerShares = key.RewardPerShares.Add(
		key.CurrentRewards.QuoDecTruncate(sdk.NewDecFromInt(key.TotalLock))...)
	key.CurrentRewards = sdk.NewDecCoins()
	k.SetKey(ctx, key)

	return key
}
