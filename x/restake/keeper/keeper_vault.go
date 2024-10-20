package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// GetOrCreateVault get the vault object by using key. If the vault doesn't exist, it will initialize the new vault.
func (k Keeper) GetOrCreateVault(ctx sdk.Context, key string) (types.Vault, error) {
	vault, found := k.GetVault(ctx, key)
	if !found {
		vault = types.NewVault(
			key,
			true,
		)

		k.SetVault(ctx, vault)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCreateVault,
				sdk.NewAttribute(types.AttributeKeyKey, key),
			),
		)
	}

	return vault, nil
}

// IsActiveVault checks whether the vault is active or not.
func (k Keeper) IsActiveVault(ctx sdk.Context, key string) bool {
	vault, found := k.GetVault(ctx, key)
	if !found {
		return false
	}

	return vault.IsActive
}

// DeactivateVault deactivates the vault.
func (k Keeper) DeactivateVault(ctx sdk.Context, key string) error {
	vault, found := k.GetVault(ctx, key)
	if !found {
		return types.ErrVaultNotFound.Wrapf("key: %s", key)
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

// -------------------------------
// store part
// -------------------------------

// GetVaultsIterator gets iterator of vault store.
func (k Keeper) GetVaultsIterator(ctx sdk.Context) storetypes.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.VaultStoreKeyPrefix)
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
	vault, found := k.GetVault(ctx, key)
	if !found {
		panic(types.ErrVaultNotFound)
	}

	return vault
}

// GetVault gets a vault from store by key.
func (k Keeper) GetVault(ctx sdk.Context, key string) (types.Vault, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.VaultStoreKey(key))
	if bz == nil {
		return types.Vault{}, false
	}

	var vault types.Vault
	k.cdc.MustUnmarshal(bz, &vault)

	return vault, true
}

// SetVault sets a vault to the store.
func (k Keeper) SetVault(ctx sdk.Context, vault types.Vault) {
	ctx.KVStore(k.storeKey).Set(types.VaultStoreKey(vault.Key), k.cdc.MustMarshal(&vault))
}
