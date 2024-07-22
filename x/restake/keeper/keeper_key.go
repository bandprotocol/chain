package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetOrCreateKey(ctx sdk.Context, keyName string) (types.Key, error) {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		keyAccAddr, err := k.createKeyAccount(ctx, keyName)
		if err != nil {
			return types.Key{}, err
		}

		key = types.Key{
			Name:            keyName,
			PoolAddress:     keyAccAddr.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardPerPowers: sdk.NewDecCoins(),
			Remainders:      sdk.NewDecCoins(),
		}

		k.SetKey(ctx, key)
	}

	return key, nil
}

// AddRewards adds rewards to the pool address and re-calculate reward_per_share of the key
func (k Keeper) AddRewards(ctx sdk.Context, sender sdk.AccAddress, keyName string, rewards sdk.Coins) error {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return err
	}

	if !key.IsActive {
		return types.ErrKeyNotActive
	}

	if key.TotalPower.IsZero() {
		return types.ErrTotalLockZero
	}

	err = k.bankKeeper.SendCoins(ctx, sender, sdk.MustAccAddressFromBech32(key.PoolAddress), rewards)
	if err != nil {
		return err
	}

	decRewards := sdk.NewDecCoinsFromCoins(rewards.Sort()...)
	totalPower := sdkmath.LegacyNewDecFromInt(key.TotalPower)
	rewardPerPowers := decRewards.QuoDecTruncate(totalPower)
	truncatedRewards := decRewards.Sub(rewardPerPowers.MulDecTruncate(totalPower))

	// add truncate part to remainder
	// e.g. rewards = 1, totalPower = 3 -> rewardPerPower = 0.333333333333333
	// remainder = 1 - (0.333333333333333 * 3)
	key.Remainders = key.Remainders.Add(truncatedRewards...)
	key.RewardPerPowers = key.RewardPerPowers.Add(rewardPerPowers...)
	k.SetKey(ctx, key)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddRewards,
			sdk.NewAttribute(types.AttributeKeyKey, keyName),
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
		),
	)

	return nil
}

func (k Keeper) IsActiveKey(ctx sdk.Context, keyName string) bool {
	key, err := k.GetKey(ctx, keyName)
	if err != nil {
		return false
	}

	return key.IsActive
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

func (k Keeper) createKeyAccount(ctx sdk.Context, key string) (sdk.AccAddress, error) {
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

// -------------------------------
// store part
// -------------------------------

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
