package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// check if after delegation is removed, the locked power is still less than total delegation
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	delegated, err := h.k.stakingKeeper.GetDelegatorBonded(sdkCtx, delAddr)
	if err != nil {
		return err
	}

	// reduce power of removing delegation from total delegation
	removingDelegation, err := h.k.stakingKeeper.GetDelegation(sdkCtx, delAddr, valAddr)
	if err != nil {
		return err
	}

	validator, err := h.k.stakingKeeper.GetValidator(sdkCtx, valAddr)
	if err != nil {
		return err
	}

	tokens := validator.TokensFromSharesTruncated(removingDelegation.Shares)
	delegated = delegated.Sub(tokens.RoundInt())

	// check if it's able to unbond
	if !h.isAbleToUnbond(sdkCtx, delAddr, delegated) {
		return types.ErrUnableToUndelegate
	}

	return nil
}

// check if after delegation is modified, the locked power is still less than total delegation
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// get total delegation
	delegated, err := h.k.stakingKeeper.GetDelegatorBonded(sdkCtx, delAddr)
	if err != nil {
		return err
	}

	// check if it's able to unbond
	if !h.isAbleToUnbond(sdkCtx, delAddr, delegated) {
		return types.ErrUnableToUndelegate
	}

	return nil
}

func (h Hooks) BeforeValidatorSlashed(_ context.Context, _ sdk.ValAddress, _ sdkmath.LegacyDec) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}

// isAbleToUnbond checks if the new total delegation is still more than locked power in the module.
func (h Hooks) isAbleToUnbond(ctx sdk.Context, addr sdk.AccAddress, delegated sdkmath.Int) bool {
	iterator := storetypes.KVStoreReversePrefixIterator(ctx.KVStore(h.k.storeKey), types.LocksByPowerIndexKey(addr))
	defer iterator.Close()

	// loop lock from high power to low power.
	for ; iterator.Valid(); iterator.Next() {
		key := string(iterator.Value())
		_, power := types.SplitLockByPowerIndexKey(iterator.Key())

		// check if the vault of lock is active.
		if h.k.IsActiveVault(ctx, key) {
			// return true if new delegation is more than or equal to locked power.
			return delegated.GTE(power)
		}
	}

	return true
}
