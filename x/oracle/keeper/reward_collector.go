package oraclekeeper

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type rewardCollector struct {
	oracleKeeper Keeper
	bankKeeper   oracletypes.BankKeeper
	collected    sdk.DecCoins
}

func (r rewardCollector) Collect(ctx sdk.Context, coins sdk.DecCoins, address sdk.AccAddress) error {
	r.collected = r.collected.Add(coins...)
	r.oracleKeeper.SetDataProviderAccumulatedReward(ctx, address, coins)
	return nil
}

func (r rewardCollector) Collected() sdk.DecCoins {
	return r.collected
}

func (r rewardCollector) CalculateReward(data []byte, pricePerByte sdk.DecCoins) sdk.DecCoins {
	return pricePerByte.MulDec(sdk.NewDecFromInt(sdk.NewInt(int64(len(data)))))
}

func newRewardCollector(oracleKeeper Keeper, bankKeeper oracletypes.BankKeeper) RewardCollector {

	return &rewardCollector{
		oracleKeeper: oracleKeeper,
		bankKeeper:   bankKeeper,
		collected:    sdk.NewDecCoins(),
	}
}
