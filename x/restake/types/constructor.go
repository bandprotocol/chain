package types

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVault(
	key string,
	vaultAddr string,
	isActive bool,
	rewardsPerPower sdk.DecCoins,
	totalPower sdkmath.Int,
	remainders sdk.DecCoins,
) Vault {
	return Vault{
		Key:             key,
		VaultAddress:    vaultAddr,
		IsActive:        isActive,
		RewardsPerPower: rewardsPerPower,
		TotalPower:      totalPower,
		Remainders:      remainders,
	}
}

func NewLock(
	stakerAddr string,
	key string,
	power sdkmath.Int,
	posRewardDebts sdk.DecCoins,
	negRewardDebts sdk.DecCoins,
) Lock {
	return Lock{
		StakerAddress:  stakerAddr,
		Key:            key,
		Power:          power,
		PosRewardDebts: posRewardDebts,
		NegRewardDebts: negRewardDebts,
	}
}

func NewReward(
	key string,
	rewards sdk.DecCoins,
) Reward {
	return Reward{
		Key:     key,
		Rewards: rewards,
	}
}
