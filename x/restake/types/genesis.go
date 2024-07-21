package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(keys []Key, locks []Lock) *GenesisState {
	return &GenesisState{
		Keys:  keys,
		Locks: locks,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Keys:  []Key{},
		Locks: []Lock{},
	}
}
