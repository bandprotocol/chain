package types

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	groupCount uint64,
	groups []Group,
	members []Member,
	signingCount uint64,
	signings []Signing,
	desGenesis []DEGenesis,
) *GenesisState {
	return &GenesisState{
		Params:       params,
		GroupCount:   groupCount,
		Groups:       groups,
		Members:      members,
		SigningCount: signingCount,
		Signings:     signings,
		DEsGenesis:   desGenesis,
	}
}

// DefaultGenesisState returns the default tss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		0,
		[]Group{},
		[]Member{},
		0,
		[]Signing{},
		[]DEGenesis{},
	)
}
