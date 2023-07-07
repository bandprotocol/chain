package types

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	groupCount uint64,
	signingCount uint64,
	groups []Group,
	deQueuesGenesis []DEQueueGenesis,
	desGenesis []DEGenesis,
) *GenesisState {
	return &GenesisState{
		Params:          params,
		GroupCount:      groupCount,
		SigningCount:    signingCount,
		Groups:          groups,
		DEQueuesGenesis: deQueuesGenesis,
		DEsGenesis:      desGenesis,
	}
}

// DefaultGenesisState returns the default tss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, []Group{}, []DEQueueGenesis{}, []DEGenesis{})
}
