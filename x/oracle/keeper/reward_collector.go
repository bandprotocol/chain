package oraclekeeper

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type rewardCollector struct {
	oracleKeeper Keeper
	bankKeeper   oracletypes.BankKeeper
	collected    sdk.Coins
}

func (r rewardCollector) Collect(ctx sdk.Context, coins sdk.Coins, address sdk.AccAddress) error {
	r.collected = r.collected.Add(coins...)
	r.oracleKeeper.SetDataProviderAccumulatedReward(ctx, address, coins)
	return nil
}

func (r rewardCollector) Collected() sdk.Coins {
	return r.collected
}

func (r rewardCollector) CalculateReward(data []byte, pricePerByte sdk.Coins) sdk.Coins {
	price := sdk.NewDecCoinsFromCoins(pricePerByte...)
	reward, _ := price.MulDec(sdk.NewDecFromInt(sdk.NewInt(int64(len(data))))).TruncateDecimal()
	return reward
}

func newRewardCollector(oracleKeeper Keeper, bankKeeper oracletypes.BankKeeper) RewardCollector {
	return &rewardCollector{
		oracleKeeper: oracleKeeper,
		bankKeeper:   bankKeeper,
		collected:    sdk.NewCoins(),
	}
}
