package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(keys []Key, stakes []Stake, rewards []RewardGenesis, remainder Remainder) *GenesisState {
	return &GenesisState{
		Keys:      keys,
		Stakes:    stakes,
		Rewards:   rewards,
		Remainder: remainder,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Keys:    []Key{},
		Stakes:  []Stake{},
		Rewards: []RewardGenesis{},
		Remainder: Remainder{
			Amounts: sdk.NewDecCoins(),
		},
	}
}
