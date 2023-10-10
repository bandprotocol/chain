package types

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	groupCount uint64,
	groups []Group,
	members []Member,
	signingCount uint64,
	signings []Signing,
	replacementCount uint64,
	replacements []Replacement,
	deQueuesGenesis []DEQueueGenesis,
	desGenesis []DEGenesis,
	statuses []Status,
) *GenesisState {
	return &GenesisState{
		Params:           params,
		GroupCount:       groupCount,
		Groups:           groups,
		Members:          members,
		SigningCount:     signingCount,
		Signings:         signings,
		ReplacementCount: replacementCount,
		Replacements:     replacements,
		DEQueuesGenesis:  deQueuesGenesis,
		DEsGenesis:       desGenesis,
		Statuses:         statuses,
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
		0,
		[]Replacement{},
		[]DEQueueGenesis{},
		[]DEGenesis{},
		[]Status{},
	)
}
