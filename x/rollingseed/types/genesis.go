package types

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState() *GenesisState {
	return &GenesisState{}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState()
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	return nil
}
