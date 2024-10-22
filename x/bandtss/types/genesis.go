package types

import "github.com/bandprotocol/chain/v3/pkg/tss"

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

// Validate performs basic validation of genesis data returning an error for
// any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}
