package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// NewGenesisState creates a new GenesisState instance.
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

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Group{},
		[]Member{},
		[]DEGenesis{},
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// validate group information.
	groupSizes := make(map[tss.GroupID]int)
	for _, group := range gs.Groups {
		if _, ok := groupSizes[group.ID]; ok {
			return fmt.Errorf("duplicate group ID %d", group.ID)
		}

		if err := group.Validate(); err != nil {
			return err
		}

		groupSizes[group.ID] = int(group.Size_)
	}

	// all members must belong to an existing group.
	memberCounts := make(map[tss.GroupID]int)
	for _, member := range gs.Members {
		if _, ok := groupSizes[member.GroupID]; !ok {
			return fmt.Errorf("invalid group ID %d for member %s", member.GroupID, member.Address)
		}

		if err := member.Validate(); err != nil {
			return err
		}

		memberCounts[member.GroupID]++
	}

	// check group size to match with numbers of existing members.
	for groupID, size := range groupSizes {
		if size != memberCounts[groupID] {
			return fmt.Errorf("group %d has %d members, expect %d", groupID, memberCounts[groupID], size)
		}
	}

	// anyone can submit DE.
	for _, de := range gs.DEs {
		if _, err := sdk.AccAddressFromBech32(de.Address); err != nil {
			return err
		}

		if err := de.DE.Validate(); err != nil {
			return err
		}
	}

	return nil
}
