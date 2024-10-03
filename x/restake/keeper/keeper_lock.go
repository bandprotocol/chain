package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

// SetLockedPower sets the new locked power of the address to the vault
// This function will override the previous locked power.
func (k Keeper) SetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string, power sdkmath.Int) error {
	if k.IsLiquidStaker(stakerAddr) {
		return types.ErrLiquidStakerNotAllowed
	}

	if !power.IsUint64() {
		return types.ErrInvalidPower
	}

	// check if total power is not less than power
	totalPower, err := k.GetTotalPower(ctx, stakerAddr)
	if err != nil {
		return err
	}

	if totalPower.LT(power) {
		return types.ErrPowerNotEnough
	}

	vault, err := k.GetOrCreateVault(ctx, key)
	if err != nil {
		return err
	}

	if !vault.IsActive {
		return types.ErrVaultNotActive
	}

	// check if there is a lock before
	lock, found := k.GetLock(ctx, stakerAddr, key)
	if !found {
		lock = types.NewLock(
			stakerAddr.String(),
			key,
			sdkmath.NewInt(0),
			sdk.NewDecCoins(),
			sdk.NewDecCoins(),
		)
	}

	diffPower := power.Sub(lock.Power)

	vault.TotalPower = vault.TotalPower.Add(diffPower)
	k.SetVault(ctx, vault)

	additionalDebts := vault.RewardsPerPower.MulDecTruncate(sdkmath.LegacyNewDecFromInt(diffPower.Abs()))
	if diffPower.IsPositive() {
		lock.PosRewardDebts = lock.PosRewardDebts.Add(additionalDebts...)
	} else {
		lock.NegRewardDebts = lock.NegRewardDebts.Add(additionalDebts...)
	}
	lock.Power = power
	k.SetLock(ctx, lock)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockPower,
			sdk.NewAttribute(types.AttributeKeyStaker, stakerAddr.String()),
			sdk.NewAttribute(types.AttributeKeyKey, key),
			sdk.NewAttribute(types.AttributeKeyPower, power.String()),
		),
	)

	return nil
}

// GetLockedPower returns locked power of the address to the vault.
func (k Keeper) GetLockedPower(ctx sdk.Context, stakerAddr sdk.AccAddress, key string) (sdkmath.Int, error) {
	if k.IsLiquidStaker(stakerAddr) {
		return sdkmath.Int{}, types.ErrLiquidStakerNotAllowed
	}

	vault, found := k.GetVault(ctx, key)
	if !found {
		return sdkmath.Int{}, types.ErrVaultNotFound.Wrapf("key: %s", key)
	}

	if !vault.IsActive {
		return sdkmath.Int{}, types.ErrVaultNotActive
	}

	lock, found := k.GetLock(ctx, stakerAddr, key)
	if !found {
		return sdkmath.Int{}, types.ErrLockNotFound.Wrapf("address: %s, key: %s", stakerAddr.String(), key)
	}

	return lock.Power, nil
}

// getAccumulatedRewards gets the accumulatedRewards of a lock if they lock since beginning.
func (k Keeper) getAccumulatedRewards(ctx sdk.Context, lock types.Lock) sdk.DecCoins {
	vault := k.MustGetVault(ctx, lock.Key)

	return vault.RewardsPerPower.MulDecTruncate(sdkmath.LegacyNewDecFromInt(lock.Power))
}

// getReward gets the reward of a lock by using accumulated rewards and reward debts.
func (k Keeper) getReward(ctx sdk.Context, lock types.Lock) types.Reward {
	totalRewards := k.getAccumulatedRewards(ctx, lock)

	return types.NewReward(
		lock.Key,
		totalRewards.Add(lock.NegRewardDebts...).Sub(lock.PosRewardDebts),
	)
}

// isValidPower checks if the new power matches with current locked power.
func (k Keeper) isValidPower(ctx sdk.Context, addr sdk.AccAddress, totalPower sdkmath.Int) bool {
	iterator := storetypes.KVStoreReversePrefixIterator(ctx.KVStore(k.storeKey), types.LocksByPowerIndexKey(addr))
	defer iterator.Close()

	// loop lock from high power to low power.
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Value())
		_, power := types.SplitLockByPowerIndexKey(iterator.Key())

		// check if the vault of lock is active.
		if k.IsActiveVault(ctx, key) {
			// return true if new delegation is more than or equal to locked power.
			return totalPower.GTE(power)
		}
	}

	return true
}

// -------------------------------
// store part
// -------------------------------

// GetLocksIterator gets iterator of lock store.
func (k Keeper) GetLocksIterator(ctx sdk.Context) storetypes.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.LockStoreKeyPrefix)
}

// GetLocksByAddressIterator gets iterator of locks of the speicfic address.
func (k Keeper) GetLocksByAddressIterator(ctx sdk.Context, addr sdk.AccAddress) storetypes.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.LocksByAddressStoreKey(addr))
}

// GetLocksByAddress gets all locks of the address.
func (k Keeper) GetLocksByAddress(ctx sdk.Context, addr sdk.AccAddress) (locks []types.Lock) {
	iterator := k.GetLocksByAddressIterator(ctx, addr)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lock types.Lock
		k.cdc.MustUnmarshal(iterator.Value(), &lock)
		locks = append(locks, lock)
	}

	return locks
}

// GetLocks gets all locks in the store.
func (k Keeper) GetLocks(ctx sdk.Context) (locks []types.Lock) {
	iterator := k.GetLocksIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lock types.Lock
		k.cdc.MustUnmarshal(iterator.Value(), &lock)
		locks = append(locks, lock)
	}

	return locks
}

// GetLock gets a lock from store by address and key.
func (k Keeper) GetLock(ctx sdk.Context, addr sdk.AccAddress, key string) (types.Lock, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.LockStoreKey(addr, key))
	if bz == nil {
		return types.Lock{}, false
	}

	var lock types.Lock
	k.cdc.MustUnmarshal(bz, &lock)

	return lock, true
}

// SetLock sets a lock to the store.
func (k Keeper) SetLock(ctx sdk.Context, lock types.Lock) {
	addr := sdk.MustAccAddressFromBech32(lock.StakerAddress)
	k.DeleteLock(ctx, addr, lock.Key)

	ctx.KVStore(k.storeKey).Set(types.LockStoreKey(addr, lock.Key), k.cdc.MustMarshal(&lock))
	k.setLockByPower(ctx, lock)
}

// DeleteLock deletes a lock from the store.
func (k Keeper) DeleteLock(ctx sdk.Context, addr sdk.AccAddress, key string) {
	lock, found := k.GetLock(ctx, addr, key)
	if !found {
		return
	}

	ctx.KVStore(k.storeKey).Delete(types.LockStoreKey(addr, key))
	k.deleteLockByPower(ctx, lock)
}

// setLockByPower sets a lock by power to the store.
func (k Keeper) setLockByPower(ctx sdk.Context, lock types.Lock) {
	ctx.KVStore(k.storeKey).Set(types.LockByPowerIndexKey(lock), []byte(lock.Key))
}

// deleteLockByPower deletes a lock by power from the store.
func (k Keeper) deleteLockByPower(ctx sdk.Context, lock types.Lock) {
	ctx.KVStore(k.storeKey).Delete(types.LockByPowerIndexKey(lock))
}
