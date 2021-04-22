package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewDataProviderAccumulatedReward(acc sdk.AccAddress, reward sdk.DecCoins) *DataProviderAccumulatedReward {
	return &DataProviderAccumulatedReward{
		DataProvider:       acc.String(),
		DataProviderReward: reward,
	}
}
