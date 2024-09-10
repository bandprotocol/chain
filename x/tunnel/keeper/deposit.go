package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// Deposit creates a new deposit for the tunnel module.
func (k Keeper) validateInitialDeposit(ctx sdk.Context, params types.Params, initialDeposit sdk.Coins) error {
	if !initialDeposit.IsValid() || initialDeposit.IsAnyNegative() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, initialDeposit.String())
	}

	minDepositCoins := k.GetParams(ctx).MinDeposit
	if !initialDeposit.IsAllGTE(minDepositCoins) {
		return errors.Wrapf(types.ErrMinDepositTooSmall, "was (%s), need (%s)", initialDeposit, minDepositCoins)
	}

	return nil
}

// validateDepositDenom validates if the deposit denom is accepted by the tunnel module.
func (k Keeper) validateDepositDenom(params types.Params, depositAmount sdk.Coins) error {
	denoms := make([]string, 0, len(params.MinDeposit))
	acceptedDenoms := make(map[string]bool, len(params.MinDeposit))
	for _, coin := range params.MinDeposit {
		acceptedDenoms[coin.Denom] = true
		denoms = append(denoms, coin.Denom)
	}

	for _, coin := range depositAmount {
		if _, ok := acceptedDenoms[coin.Denom]; !ok {
			return errors.Wrapf(
				types.ErrInvalidDepositDenom,
				"deposited %s, but gov accepts only the following denom(s): %v",
				depositAmount,
				denoms,
			)
		}
	}

	return nil
}
