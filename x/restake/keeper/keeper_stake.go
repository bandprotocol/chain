package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetStakesIterator(ctx sdk.Context, address sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StakesStoreKey(address))
}

func (k Keeper) GetActiveStakes(ctx sdk.Context, address sdk.AccAddress) (stakes []types.Stake) {
	iterator := k.GetStakesIterator(ctx, address)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.Stake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)

		if !k.IsActiveKey(ctx, stake.Key) {
			continue
		}

		stakes = append(stakes, stake)
	}

	return stakes
}

func (k Keeper) GetStakes(ctx sdk.Context, address sdk.AccAddress) (stakes []types.Stake) {
	iterator := k.GetStakesIterator(ctx, address)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.Stake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}

	return stakes
}

func (k Keeper) HasStake(ctx sdk.Context, address sdk.AccAddress, keyName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.StakeStoreKey(address, keyName))
}

func (k Keeper) GetStake(ctx sdk.Context, address sdk.AccAddress, keyName string) (types.Stake, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.StakeStoreKey(address, keyName))
	if bz == nil {
		return types.Stake{}, types.ErrStakeNotFound.Wrapf(
			"failed to get stake of %s with key name: %s",
			address.String(),
			keyName,
		)
	}

	var stake types.Stake
	k.cdc.MustUnmarshal(bz, &stake)

	return stake, nil
}

func (k Keeper) SetStake(ctx sdk.Context, stake types.Stake) {
	address := sdk.MustAccAddressFromBech32(stake.Address)
	k.DeleteStake(ctx, address, stake.Key)

	ctx.KVStore(k.storeKey).Set(types.StakeStoreKey(address, stake.Key), k.cdc.MustMarshal(&stake))
	k.setStakeByAmount(ctx, stake)
}

func (k Keeper) setStakeByAmount(ctx sdk.Context, stake types.Stake) {
	ctx.KVStore(k.storeKey).Set(types.StakeByAmountIndexKey(stake), []byte(stake.Key))
}

func (k Keeper) DeleteStake(ctx sdk.Context, address sdk.AccAddress, keyName string) {
	stake, err := k.GetStake(ctx, address, keyName)
	if err != nil {
		return
	}
	ctx.KVStore(k.storeKey).Delete(types.StakeStoreKey(address, keyName))
	k.deleteStakeByAmount(ctx, stake)
}

func (k Keeper) deleteStakeByAmount(ctx sdk.Context, stake types.Stake) {
	ctx.KVStore(k.storeKey).Delete(types.StakeByAmountIndexKey(stake))
}

func (k Keeper) ProcessStake(ctx sdk.Context, stake types.Stake) {
	key := k.MustGetKey(ctx, stake.Key)
	key = k.ProcessKey(ctx, key)

	address := sdk.MustAccAddressFromBech32(stake.Address)
	diff := key.RewardPerShares.Sub(stake.RewardDebts)

	// update reward
	if !diff.IsZero() {
		reward := k.GetOrCreateReward(ctx, address, stake.Key)
		reward.Amounts = reward.Amounts.Add(diff.MulDecTruncate(sdk.NewDecFromInt(stake.Amount))...)
		k.SetReward(ctx, address, reward)

		// update debt
		stake.RewardDebts = key.RewardPerShares
		k.SetStake(ctx, stake)
	}

	if !key.IsActive {
		k.DeleteStake(ctx, address, stake.Key)
	}
}
