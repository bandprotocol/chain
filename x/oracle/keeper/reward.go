package keeper

import (
	"github.com/GeoDB-Limited/odincore/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
)

func (k Keeper) SetDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress, reward sdk.DecCoin) {
	key := types.DataProviderRewardsPrefixKey(acc)
	if !k.HasDataProviderReward(ctx, acc) {
		ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshalBinaryBare(reward))
		return
	}
	oldReward := k.GetDataProviderAccumulatedReward(ctx, acc)
	newReward := oldReward.Add(reward)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshalBinaryBare(newReward))
}

func (k Keeper) ClearDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(types.DataProviderRewardsPrefixKey(acc))
}

func (k Keeper) GetDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress) (reward sdk.DecCoin) {
	key := types.DataProviderRewardsPrefixKey(acc)
	bz := ctx.KVStore(k.storeKey).Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &reward)
	return reward
}

func (k Keeper) HasDataProviderReward(ctx sdk.Context, acc sdk.AccAddress) bool {
	return ctx.KVStore(k.storeKey).Has(types.DataProviderRewardsPrefixKey(acc))
}

// sends rewards from fee pool to data providers, that have given data for the passed request
func (k Keeper) AllocateRewardsToDataProviders(ctx sdk.Context, rid types.RequestID) {
	logger := k.Logger(ctx)
	request := k.MustGetRequest(ctx, rid)

	// rewards are lying in the distribution fee pool
	feePool := k.distrKeeper.GetFeePool(ctx)

	for _, rawReq := range request.RawRequests {
		ds := k.MustGetDataSource(ctx, rawReq.GetDataSourceID())
		if !k.HasDataProviderReward(ctx, ds.Owner) {
			continue
		}
		reward := k.GetDataProviderAccumulatedReward(ctx, ds.Owner)

		diff, hasNeg := feePool.CommunityPool.SafeSub(sdk.NewDecCoins(reward))
		if hasNeg {
			logger.With("lack", diff, "denom", reward.Denom).Error("oracle pool does not have enough coins to reward data providers")
			// not return because maybe still enough coins to pay someone
			continue
		}
		feePool.CommunityPool = diff

		rewardCoin, remainder := reward.TruncateDecimal()
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, distr.ModuleName, ds.Owner, sdk.NewCoins(rewardCoin))
		if err != nil {
			panic(err)
		}

		// we are sure to have paid the reward to the provider, we can remove him now
		k.ClearDataProviderAccumulatedReward(ctx, ds.Owner)

		// if there is something left, that we cannot pay now, we can store it for later
		if remainder.IsPositive() {
			k.SetDataProviderAccumulatedReward(ctx, ds.Owner, remainder)
		}
	}

	k.distrKeeper.SetFeePool(ctx, feePool)
}
