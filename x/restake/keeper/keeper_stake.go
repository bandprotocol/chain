package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// GetStakedPower returns the power from staked coins in the module.
func (k Keeper) GetStakedPower(ctx sdk.Context, stakerAddr sdk.AccAddress) sdkmath.Int {
	stake := k.GetStake(ctx, stakerAddr)

	power := sdkmath.NewInt(0)
	allowedDenoms := k.GetParams(ctx).AllowedDenoms
	for _, denom := range allowedDenoms {
		power = power.Add(stake.Coins.AmountOf(denom))
	}

	return power
}

// GetStakesIterator gets iterator of stake store.
func (k Keeper) GetStakesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StakeStoreKeyPrefix)
}

// GetStakes gets all stakes in the store.
func (k Keeper) GetStakes(ctx sdk.Context) (stakes []types.Stake) {
	iterator := k.GetStakesIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.Stake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}

	return stakes
}

// GetStake gets a stake from store by address and denom.
func (k Keeper) GetStake(ctx sdk.Context, addr sdk.AccAddress) types.Stake {
	bz := ctx.KVStore(k.storeKey).Get(types.StakeStoreKey(addr))
	if bz == nil {
		return types.NewStake(
			addr.String(),
			sdk.NewCoins(),
		)
	}

	var stake types.Stake
	k.cdc.MustUnmarshal(bz, &stake)

	return stake
}

// SetStake sets a stake to the store.
func (k Keeper) SetStake(ctx sdk.Context, stake types.Stake) {
	addr := sdk.MustAccAddressFromBech32(stake.StakerAddress)

	if stake.Coins.IsZero() {
		k.DeleteStake(ctx, addr)
		return
	}

	ctx.KVStore(k.storeKey).Set(types.StakeStoreKey(addr), k.cdc.MustMarshal(&stake))
}

// DeleteStake deletes a stake from the store.
func (k Keeper) DeleteStake(ctx sdk.Context, addr sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(types.StakeStoreKey(addr))
}
