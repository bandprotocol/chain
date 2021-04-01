package oraclekeeper

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (k Keeper) SetDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress, reward sdk.DecCoin) {
	key := oracletypes.DataProviderRewardsPrefixKey(acc)
	if !k.HasDataProviderReward(ctx, acc) {
		ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshalBinaryBare(&reward))
		return
	}
	oldReward := k.GetDataProviderAccumulatedReward(ctx, acc)
	newReward := oldReward.Add(reward)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshalBinaryBare(&newReward))
}

func (k Keeper) ClearDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress) {
	ctx.KVStore(k.storeKey).Delete(oracletypes.DataProviderRewardsPrefixKey(acc))
}

func (k Keeper) GetDataProviderAccumulatedReward(ctx sdk.Context, acc sdk.AccAddress) (reward sdk.DecCoin) {
	key := oracletypes.DataProviderRewardsPrefixKey(acc)
	bz := ctx.KVStore(k.storeKey).Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &reward)
	return reward
}

func (k Keeper) HasDataProviderReward(ctx sdk.Context, acc sdk.AccAddress) bool {
	return ctx.KVStore(k.storeKey).Has(oracletypes.DataProviderRewardsPrefixKey(acc))
}

// sends rewards from fee pool to data providers, that have given data for the passed request
func (k Keeper) AllocateRewardsToDataProviders(ctx sdk.Context, rid oracletypes.RequestID) {
	logger := k.Logger(ctx)
	request := k.MustGetRequest(ctx, rid)

	// rewards are lying in the distribution fee pool
	feePool := k.distrKeeper.GetFeePool(ctx)

	for _, rawReq := range request.RawRequests {
		ds := k.MustGetDataSource(ctx, rawReq.GetDataSourceID())

		ownerAccAddr, err := sdk.AccAddressFromBech32(ds.Owner)
		if err != nil {
			panic(err)
		}
		if !k.HasDataProviderReward(ctx, ownerAccAddr) {
			continue
		}
		reward := k.GetDataProviderAccumulatedReward(ctx, ownerAccAddr)

		diff, hasNeg := feePool.CommunityPool.SafeSub(sdk.NewDecCoins(reward))
		if hasNeg {
			logger.With("lack", diff, "denom", reward.Denom).Error("oracle pool does not have enough coins to reward data providers")
			// not return because maybe still enough coins to pay someone
			continue
		}
		feePool.CommunityPool = diff

		rewardCoin, remainder := reward.TruncateDecimal()
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, distrtypes.ModuleName, ownerAccAddr, sdk.NewCoins(rewardCoin))
		if err != nil {
			panic(err)
		}

		// we are sure to have paid the reward to the provider, we can remove him now
		k.ClearDataProviderAccumulatedReward(ctx, ownerAccAddr)

		// if there is something left, that we cannot pay now, we can store it for later
		if remainder.IsPositive() {
			k.SetDataProviderAccumulatedReward(ctx, ownerAccAddr, remainder)
		}
	}

	k.distrKeeper.SetFeePool(ctx, feePool)
}
