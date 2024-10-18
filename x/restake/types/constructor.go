package types

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewVault(
	key string,
	isActive bool,
) Vault {
	return Vault{
		Key:      key,
		IsActive: isActive,
	}
}

func NewLock(
	stakerAddr string,
	key string,
	power sdkmath.Int,
) Lock {
	return Lock{
		StakerAddress: stakerAddr,
		Key:           key,
		Power:         power,
	}
}

func NewStake(
	stakerAddr string,
	coins sdk.Coins,
) Stake {
	return Stake{
		StakerAddress: stakerAddr,
		Coins:         coins,
	}
}
