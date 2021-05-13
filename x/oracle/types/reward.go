package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewDataProviderAccumulatedReward(acc sdk.AccAddress, reward sdk.Coins) *DataProviderAccumulatedReward {
	return &DataProviderAccumulatedReward{
		DataProvider:       acc.String(),
		DataProviderReward: reward,
	}
}

func NewDataProvidersAccumulatedRewards(reward sdk.Coins) DataProvidersAccumulatedRewards {
	return DataProvidersAccumulatedRewards{
		Amount: reward,
	}
}
