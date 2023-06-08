package types

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState returns the default tss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams())
}
