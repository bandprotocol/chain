package types

import (
	"fmt"
)

// NewGenesisState creates a new GenesisState instance.
func NewGenesisState(
	params Params,
	members []Member,
	currentGroup CurrentGroup,
) *GenesisState {
	return &GenesisState{
		Params:       params,
		Members:      members,
		CurrentGroup: currentGroup,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Member{},
		CurrentGroup{},
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, m := range gs.Members {
		if err := m.Validate(); err != nil {
			return err
		}

		if m.GroupID != gs.CurrentGroup.GroupID {
			return fmt.Errorf("member %s is not in current group %d", m.Address, gs.CurrentGroup.GroupID)
		}
	}

	return nil
}
