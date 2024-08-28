package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterValidatorCreated(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeValidatorModified(_ sdk.Context, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorRemoved(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBonded(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterValidatorBeginUnbonding(_ sdk.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationCreated(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h Hooks) BeforeDelegationSharesModified(_ sdk.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

// check if after delegation is removed, the locked power is still less than total delegation
func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	delegated := h.k.stakingKeeper.GetDelegatorBonded(ctx, delAddr)

	// reduce power of removing delegation from total delegation
	removingDelegation, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if found {
		validatorAddr, err := sdk.ValAddressFromBech32(removingDelegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := h.k.stakingKeeper.GetValidator(ctx, validatorAddr)
		if found {
			shares := removingDelegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			delegated = delegated.Sub(tokens.RoundInt())
		}
	}

	// check if it's able to unbond
	if !h.isAbleToUnbond(ctx, delAddr, delegated) {
		return types.ErrUnableToUndelegate
	}

	return nil
}

// check if after delegation is modified, the locked power is still less than total delegation
func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	// get total delegation
	delegated := h.k.stakingKeeper.GetDelegatorBonded(ctx, delAddr)

	// check if it's able to unbond
	if !h.isAbleToUnbond(ctx, delAddr, delegated) {
		return types.ErrUnableToUndelegate
	}

	return nil
}

func (h Hooks) BeforeValidatorSlashed(_ sdk.Context, _ sdk.ValAddress, _ sdk.Dec) error {
	return nil
}

func (h Hooks) AfterUnbondingInitiated(_ sdk.Context, _ uint64) error {
	return nil
}

// isAbleToUnbond checks if the new total delegation is still more than locked power in the module.
func (h Hooks) isAbleToUnbond(ctx sdk.Context, addr sdk.AccAddress, delegated sdkmath.Int) bool {
	stakedPower := h.k.GetStakedPower(ctx, addr)
	totalPower := stakedPower.Add(delegated)

	return h.k.isValidPower(ctx, addr, totalPower)
}
