package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (k Keeper) GetOrCreateReward(ctx sdk.Context, address sdk.AccAddress, keyName string) types.Reward {
	reward, err := k.GetReward(ctx, address, keyName)
	if err != nil {
		reward = types.Reward{
			Key:     keyName,
			Amounts: sdk.NewDecCoins(),
		}

		k.SetReward(ctx, address, reward)
	}

	return reward
}

func (k Keeper) GetRewardsIterator(ctx sdk.Context, address sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.RewardsStoreKey(address))
}

func (k Keeper) GetRewards(ctx sdk.Context, address sdk.AccAddress) (rewards []types.Reward) {
	iterator := k.GetRewardsIterator(ctx, address)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var reward types.Reward
		k.cdc.MustUnmarshal(iterator.Value(), &reward)
		rewards = append(rewards, reward)
	}

	return rewards
}

func (k Keeper) HasReward(ctx sdk.Context, address sdk.AccAddress, keyName string) bool {
	return ctx.KVStore(k.storeKey).Has(types.RewardStoreKey(address, keyName))
}

func (k Keeper) GetReward(ctx sdk.Context, address sdk.AccAddress, keyName string) (types.Reward, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.RewardStoreKey(address, keyName))
	if bz == nil {
		return types.Reward{}, types.ErrRewardNotFound.Wrapf(
			"failed to get reward of %s with key name: %s",
			address.String(),
			keyName,
		)
	}

	var reward types.Reward
	k.cdc.MustUnmarshal(bz, &reward)

	return reward, nil
}

func (k Keeper) SetReward(ctx sdk.Context, address sdk.AccAddress, reward types.Reward) {
	ctx.KVStore(k.storeKey).Set(types.RewardStoreKey(address, reward.Key), k.cdc.MustMarshal(&reward))
}

func (k Keeper) DeleteReward(ctx sdk.Context, address sdk.AccAddress, keyName string) {
	ctx.KVStore(k.storeKey).Delete(types.RewardStoreKey(address, keyName))
}
