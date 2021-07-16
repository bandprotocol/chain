package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewDataProviderAccumulatedReward(acc sdk.AccAddress, reward sdk.Coins) *DataProviderAccumulatedReward {
	return &DataProviderAccumulatedReward{
		DataProvider:       acc.String(),
		DataProviderReward: reward,
	}
}

func NewDataProvidersAccumulatedRewards(currentRewardPerByte sdk.Coins, accumulatedAmount sdk.Coins) DataProvidersAccumulatedRewards {
	return DataProvidersAccumulatedRewards{
		CurrentRewardPerByte: currentRewardPerByte,
		AccumulatedAmount:    accumulatedAmount,
	}
}
