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
	groupSizes := make(map[tss.GroupID]uint64)
	for _, group := range gs.Groups {
		if _, ok := groupSizes[group.ID]; ok {
			return fmt.Errorf("duplicate group ID %d", group.ID)
		}

		if err := group.Validate(); err != nil {
			return err
		}

		groupSizes[group.ID] = group.Size_
	}

	// all members must belong to an existing group.
	memberCounts := make(map[tss.GroupID]uint64)
	seenMemberIDGroups := make(map[string]bool)
	seenMemberAddressGroups := make(map[string]bool)
	for _, member := range gs.Members {
		if err := member.Validate(); err != nil {
			return err
		}

		if size, ok := groupSizes[member.GroupID]; !ok || uint64(member.ID) > size {
			return fmt.Errorf("invalid group ID %d for member %d", member.GroupID, member.ID)
		}

		// validate duplicate member ID in the same group.
		memberIDGroupKey := fmt.Sprintf("%d-%d", member.ID, member.GroupID)
		if seenMemberIDGroups[memberIDGroupKey] {
			return fmt.Errorf("duplicate member ID %d in group ID %d", member.ID, member.GroupID)
		}
		seenMemberIDGroups[memberIDGroupKey] = true

		// validate duplicate member address in the same group.
		memberAddressGroupKey := fmt.Sprintf("%s-%d", member.Address, member.GroupID)
		if seenMemberAddressGroups[memberAddressGroupKey] {
			return fmt.Errorf("duplicate member Address %s in group ID %d", member.Address, member.GroupID)
		}
		seenMemberAddressGroups[memberAddressGroupKey] = true

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
