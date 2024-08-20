package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// GetOrCreateVault get the vault object by using key. If the vault doesn't exist, it will initialize the new vault.
func (k Keeper) GetOrCreateVault(ctx sdk.Context, key string) (types.Vault, error) {
	vault, err := k.GetVault(ctx, key)
	if err != nil {
		vaultAccAddr, err := k.createVaultAccount(ctx, key)
		if err != nil {
			return types.Vault{}, err
		}

		vault = types.Vault{
			Key:             key,
			VaultAddress:    vaultAccAddr.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardsPerPower: sdk.NewDecCoins(),
			Remainders:      sdk.NewDecCoins(),
		}

		k.SetVault(ctx, vault)
	}

	return vault, nil
}

// AddRewards adds rewards to the pool address and re-calculate `RewardsPerPower` and `remainders` of the vault
func (k Keeper) AddRewards(ctx sdk.Context, sender sdk.AccAddress, key string, rewards sdk.Coins) error {
	vault, err := k.GetVault(ctx, key)
	if err != nil {
		return err
	}

	if !vault.IsActive {
		return types.ErrVaultNotActive
	}

	if vault.TotalPower.IsZero() {
		return types.ErrTotalPowerZero
	}

	err = k.bankKeeper.SendCoins(ctx, sender, sdk.MustAccAddressFromBech32(vault.VaultAddress), rewards)
	if err != nil {
		return err
	}

	decRewards := sdk.NewDecCoinsFromCoins(rewards.Sort()...)
	totalPower := sdkmath.LegacyNewDecFromInt(vault.TotalPower)
	RewardsPerPower := decRewards.QuoDecTruncate(totalPower)
	truncatedRewards := decRewards.Sub(RewardsPerPower.MulDecTruncate(totalPower))

	// add truncate part to remainder
	// e.g. rewards = 1, totalPower = 3 -> rewardsPerPower = 0.333333333333333
	// remainder = 1 - (0.333333333333333 * 3)
	vault.Remainders = vault.Remainders.Add(truncatedRewards...)
	vault.RewardsPerPower = vault.RewardsPerPower.Add(RewardsPerPower...)
	k.SetVault(ctx, vault)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddRewards,
			sdk.NewAttribute(types.AttributeKeyKey, key),
			sdk.NewAttribute(types.AttributeKeyRewards, rewards.String()),
		),
	)

	return nil
}

// IsActiveVault checks whether the vault is active or not.
func (k Keeper) IsActiveVault(ctx sdk.Context, key string) bool {
	vault, err := k.GetVault(ctx, key)
	if err != nil {
		return false
	}

	return vault.IsActive
}

// DeactivateVault deactivates the vault.
func (k Keeper) DeactivateVault(ctx sdk.Context, key string) error {
	vault, err := k.GetVault(ctx, key)
	if err != nil {
		return err
	}

	if !vault.IsActive {
		return types.ErrVaultAlreadyDeactivated
	}

	vault.IsActive = false
	k.SetVault(ctx, vault)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeactivateVault,
			sdk.NewAttribute(types.AttributeKeyKey, key),
		),
	)

	return nil
}

// createVaultAccount creates a vault account by using name and block hash.
func (k Keeper) createVaultAccount(ctx sdk.Context, key string) (sdk.AccAddress, error) {
	header := ctx.BlockHeader()

	buf := []byte(key)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	moduleCred, err := authtypes.NewModuleCredential(types.ModuleName, []byte(types.VaultAccountsKey), buf)
	if err != nil {
		return nil, err
	}

	vaultAccAddr := sdk.AccAddress(moduleCred.Address())

	// This should not happen
	if acc := k.authKeeper.GetAccount(ctx, vaultAccAddr); acc != nil {
		return nil, types.ErrAccountAlreadyExist.Wrapf(
			"existing account for newly generated vault account address %s",
			vaultAccAddr.String(),
		)
	}

	vaultAcc, err := authtypes.NewBaseAccountWithPubKey(moduleCred)
	if err != nil {
		return nil, err
	}

	k.authKeeper.NewAccount(ctx, vaultAcc)
	k.authKeeper.SetAccount(ctx, vaultAcc)

	return vaultAccAddr, nil
}

// -------------------------------
// store part
// -------------------------------

// GetVaultsIterator gets iterator of vault store.
func (k Keeper) GetVaultsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.VaultStoreKeyPrefix)
}

// GetVaults gets all vaults in the store.
func (k Keeper) GetVaults(ctx sdk.Context) (vaults []types.Vault) {
	iterator := k.GetVaultsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vault types.Vault
		k.cdc.MustUnmarshal(iterator.Value(), &vault)
		vaults = append(vaults, vault)
	}

	return vaults
}

// HasVault checks if vault exists in the store.
func (k Keeper) HasVault(ctx sdk.Context, vaultName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.VaultStoreKey(vaultName))
}

// MustGetVault gets a vault from store by name.
// Panics if can't get the vault.
func (k Keeper) MustGetVault(ctx sdk.Context, key string) types.Vault {
	vault, err := k.GetVault(ctx, key)
	if err != nil {
		panic(err)
	}

	return vault
}

// GetVault gets a vault from store by key.
func (k Keeper) GetVault(ctx sdk.Context, key string) (types.Vault, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.VaultStoreKey(key))
	if bz == nil {
		return types.Vault{}, types.ErrVaultNotFound.Wrapf("failed to get vault with name: %s", key)
	}

	var vault types.Vault
	k.cdc.MustUnmarshal(bz, &vault)

	return vault, nil
}

// SetVault sets a vault to the store.
func (k Keeper) SetVault(ctx sdk.Context, vault types.Vault) {
	ctx.KVStore(k.storeKey).Set(types.VaultStoreKey(vault.Key), k.cdc.MustMarshal(&vault))
}
