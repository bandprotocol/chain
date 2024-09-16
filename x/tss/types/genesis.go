package types

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	groups []Group,
	members []Member,
	desGenesis []DEGenesis,
) *GenesisState {
	return &GenesisState{
		Params:  params,
		Groups:  groups,
		Members: members,
		DEs:     desGenesis,
	}
}

// DefaultGenesisState returns the default tss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Group{},
		[]Member{},
		[]DEGenesis{},
	)
}
