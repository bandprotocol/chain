package types

// NewGenesisState - Create a new genesis state
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// DefaultGenesisState returns the default tss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState()
}
