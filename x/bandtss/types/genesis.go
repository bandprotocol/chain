package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	members []Member,
	currentGroupID tss.GroupID,
	signingCount uint64,
	signings []Signing,
	signingIDMappings []SigningIDMappingGenesis,
	replacement Replacement,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		Members:           members,
		CurrentGroupID:    currentGroupID,
		SigningCount:      signingCount,
		Signings:          signings,
		SigningIDMappings: signingIDMappings,
		Replacement:       replacement,
	}
}

// DefaultGenesisState returns the default bandtss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Member{},
		0,
		0,
		[]Signing{},
		[]SigningIDMappingGenesis{},
		Replacement{},
	)
}
