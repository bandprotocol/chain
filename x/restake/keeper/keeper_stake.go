package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetStakesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StakeStoreKeyPrefix)
}

func (k Keeper) GetStakesByAddressIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.StakesStoreKey(addr))
}

func (k Keeper) GetActiveStakes(ctx sdk.Context, addr sdk.AccAddress) (stakes []types.Stake) {
	iterator := k.GetStakesByAddressIterator(ctx, addr)
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

func (k Keeper) GetStakesByAddress(ctx sdk.Context, addr sdk.AccAddress) (stakes []types.Stake) {
	iterator := k.GetStakesByAddressIterator(ctx, addr)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.Stake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}

	return stakes
}

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

func (k Keeper) HasStake(ctx sdk.Context, addr sdk.AccAddress, keyName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.StakeStoreKey(addr, keyName))
}

func (k Keeper) GetStake(ctx sdk.Context, addr sdk.AccAddress, keyName string) (types.Stake, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.StakeStoreKey(addr, keyName))
	if bz == nil {
		return types.Stake{}, types.ErrStakeNotFound.Wrapf(
			"failed to get stake of %s with key name: %s",
			addr.String(),
			keyName,
		)
	}

	var stake types.Stake
	k.cdc.MustUnmarshal(bz, &stake)

	return stake, nil
}

func (k Keeper) SetStake(ctx sdk.Context, stake types.Stake) {
	addr := sdk.MustAccAddressFromBech32(stake.StakerAddress)
	k.DeleteStake(ctx, addr, stake.Key)

	ctx.KVStore(k.storeKey).Set(types.StakeStoreKey(addr, stake.Key), k.cdc.MustMarshal(&stake))
	k.setStakeByAmount(ctx, stake)
}

func (k Keeper) DeleteStake(ctx sdk.Context, addr sdk.AccAddress, keyName string) {
	stake, err := k.GetStake(ctx, addr, keyName)
	if err != nil {
		return
	}
	ctx.KVStore(k.storeKey).Delete(types.StakeStoreKey(addr, keyName))
	k.deleteStakeByAmount(ctx, stake)
}

func (k Keeper) setStakeByAmount(ctx sdk.Context, stake types.Stake) {
	ctx.KVStore(k.storeKey).Set(types.StakeByAmountIndexKey(stake), []byte(stake.Key))
}

func (k Keeper) deleteStakeByAmount(ctx sdk.Context, stake types.Stake) {
	ctx.KVStore(k.storeKey).Delete(types.StakeByAmountIndexKey(stake))
}

func (k Keeper) getTotalRewards(ctx sdk.Context, stake types.Stake) sdk.DecCoins {
	key := k.MustGetKey(ctx, stake.Key)

	return key.RewardPerShares.MulDecTruncate(sdk.NewDecFromInt(stake.Amount))
}

func (k Keeper) getReward(ctx sdk.Context, stake types.Stake) types.Reward {
	totalRewards := k.getTotalRewards(ctx, stake)

	return types.Reward{
		Key: stake.Key,
		Rewards: totalRewards.Add(sdk.NewDecCoinsFromCoins(stake.NegRewardDebts...)...).
			Sub(sdk.NewDecCoinsFromCoins(stake.PosRewardDebts...)),
	}
}
