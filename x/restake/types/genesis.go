package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(keys []Key, stakes []Stake) *GenesisState {
	return &GenesisState{
		Keys:   keys,
		Stakes: stakes,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Keys:   []Key{},
		Stakes: []Stake{},
	}
}
