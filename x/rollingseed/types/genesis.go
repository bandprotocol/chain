package types

// NewGenesisState - Create a new genesis state
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// DefaultGenesisState returns the default rollingseed genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState()
}

func (gs GenesisState) Validate() error {
	return nil
}
