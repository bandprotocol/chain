package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(keys []Key, stakes []Stake, rewards []RewardGenesis) *GenesisState {
	return &GenesisState{
		Keys:    keys,
		Stakes:  stakes,
		Rewards: rewards,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Keys:    []Key{},
		Stakes:  []Stake{},
		Rewards: []RewardGenesis{},
	}
}
