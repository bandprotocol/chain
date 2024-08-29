package types

import "github.com/bandprotocol/chain/v2/pkg/tss"

// NewGenesisState - Create a new genesis state
func NewGenesisState(
	params Params,
	members []Member,
	currentGroupID tss.GroupID,
) *GenesisState {
	return &GenesisState{
		Params:         params,
		Members:        members,
		CurrentGroupID: currentGroupID,
	}
}

// DefaultGenesisState returns the default bandtss genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Member{},
		0,
	)
}
