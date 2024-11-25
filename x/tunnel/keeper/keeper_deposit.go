package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// DepositToTunnel deposits to a tunnel
func (k Keeper) DepositToTunnel(
	ctx sdk.Context,
	tunnelID uint64,
	depositor sdk.AccAddress,
	depositAmount sdk.Coins,
) error {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	if err := k.validateDepositDenom(ctx, depositAmount); err != nil {
		return err
	}

	// transfer the deposit from the depositor to the tunnel module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		depositor,
		types.ModuleName,
		depositAmount,
	); err != nil {
		return err
	}

	// update the depositor's deposit
	deposit, found := k.GetDeposit(ctx, tunnelID, depositor)
	if !found {
		deposit = types.NewDeposit(tunnelID, depositor.String(), depositAmount)
	} else {
		deposit.Amount = deposit.Amount.Add(depositAmount...)
	}
	k.SetDeposit(ctx, deposit)

	// update the tunnel's total deposit
	tunnel.TotalDeposit = tunnel.TotalDeposit.Add(depositAmount...)
	k.SetTunnel(ctx, tunnel)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDepositToTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
		sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, depositAmount.String()),
	))

	return nil
}

// SetDeposit sets a deposit in the store
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit) {
	ctx.KVStore(k.storeKey).
		Set(types.DepositStoreKey(deposit.TunnelID, sdk.MustAccAddressFromBech32(deposit.Depositor)), k.cdc.MustMarshal(&deposit))
}

// GetDeposit retrieves a deposit by its tunnel ID and depositor address
func (k Keeper) GetDeposit(
	ctx sdk.Context,
	tunnelID uint64,
	depositorAddr sdk.AccAddress,
) (deposit types.Deposit, found bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.DepositStoreKey(tunnelID, depositorAddr))
	if bz == nil {
		return types.Deposit{}, false
	}

	k.cdc.MustUnmarshal(bz, &deposit)
	return deposit, true
}

// GetDeposits retrieves all deposits for the tunnel
func (k Keeper) GetDeposits(ctx sdk.Context, tunnelID uint64) []types.Deposit {
	var deposits []types.Deposit
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DepositsStoreKey(tunnelID))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// GetAllDeposits returns all deposits in the store
func (k Keeper) GetAllDeposits(ctx sdk.Context) []types.Deposit {
	var deposits []types.Deposit
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DepositStoreKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		k.cdc.MustUnmarshal(iterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	return deposits
}

// DeleteDeposit deletes a deposit from the store
func (k Keeper) DeleteDeposit(ctx sdk.Context, tunnelID uint64, depositorAddr sdk.AccAddress) {
	ctx.KVStore(k.storeKey).
		Delete(types.DepositStoreKey(tunnelID, depositorAddr))
}

// WithdrawFromTunnel withdraws a deposit from a tunnel
func (k Keeper) WithdrawFromTunnel(
	ctx sdk.Context,
	tunnelID uint64,
	amount sdk.Coins,
	withdrawer sdk.AccAddress,
) error {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	deposit, found := k.GetDeposit(ctx, tunnelID, withdrawer)
	if !found {
		return types.ErrDepositNotFound
	}

	// check if the withdrawer has enough deposit
	if !deposit.Amount.IsAllGTE(amount) {
		return types.ErrInsufficientDeposit
	}

	// transfer the deposit from the tunnel module account to the withdrawer
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		withdrawer,
		amount,
	); err != nil {
		return err
	}

	// update the withdrawer's deposit
	deposit.Amount = deposit.Amount.Sub(amount...)
	if deposit.Amount.IsZero() {
		k.DeleteDeposit(ctx, tunnelID, withdrawer)
	} else {
		k.SetDeposit(ctx, deposit)
	}

	// update the tunnel's total deposit
	tunnel.TotalDeposit = tunnel.TotalDeposit.Sub(amount...)
	k.SetTunnel(ctx, tunnel)

	// deactivate the tunnel if the total deposit is less than the min deposit
	minDeposit := k.GetParams(ctx).MinDeposit
	if tunnel.IsActive && !tunnel.TotalDeposit.IsAllGTE(minDeposit) {
		// deactivate the tunnel if the total deposit is less than the min deposit
		// error should not happen here since the tunnel is already validated
		err := k.DeactivateTunnel(ctx, tunnelID)
		if err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeWithdrawFromTunnel,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
		sdk.NewAttribute(types.AttributeKeyWithdrawer, withdrawer.String()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))

	return nil
}

// validateDepositDenom validates if the deposit denom is accepted by the tunnel module.
func (k Keeper) validateDepositDenom(ctx sdk.Context, depositAmount sdk.Coins) error {
	params := k.GetParams(ctx)

	denoms := make([]string, 0, len(params.MinDeposit))
	acceptedDenoms := make(map[string]bool, len(params.MinDeposit))
	for _, coin := range params.MinDeposit {
		acceptedDenoms[coin.Denom] = true
		denoms = append(denoms, coin.Denom)
	}

	for _, coin := range depositAmount {
		if _, ok := acceptedDenoms[coin.Denom]; !ok {
			return types.ErrInvalidDepositDenom.Wrapf(
				"deposited %s, but tunnel accepts only the following denom(s): %v",
				depositAmount,
				denoms,
			)
		}
	}

	return nil
}
