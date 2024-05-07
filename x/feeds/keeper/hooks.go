package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
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

func (h Hooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	delegated := h.k.stakingKeeper.GetDelegatorBonded(ctx, delAddr)

	// remove power of removed delegation from total delegation
	removedDelegation, found := h.k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if found {
		validatorAddr, err := sdk.ValAddressFromBech32(removedDelegation.ValidatorAddress)
		if err != nil {
			panic(err) // shouldn't happen
		}
		validator, found := h.k.stakingKeeper.GetValidator(ctx, validatorAddr)
		if found {
			shares := removedDelegation.Shares
			tokens := validator.TokensFromSharesTruncated(shares)
			delegated = delegated.Sub(tokens.RoundInt())
		}
	}

	power := sumPower(h.k.GetDelegatorSignals(ctx, delAddr))
	if power > delegated.Int64() {
		return types.ErrUnableToUndelegate
	}
	return nil
}

func (h Hooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, _ sdk.ValAddress) error {
	delegated := h.k.stakingKeeper.GetDelegatorBonded(ctx, delAddr).Int64()
	power := sumPower(h.k.GetDelegatorSignals(ctx, delAddr))
	if power > delegated {
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
