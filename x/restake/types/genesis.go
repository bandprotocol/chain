package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, vaults []Vault, locks []Lock, stakes []Stake) *GenesisState {
	return &GenesisState{
		Params: params,
		Vaults: vaults,
		Locks:  locks,
		Stakes: stakes,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultParams(),
		[]Vault{},
		[]Lock{},
		[]Stake{},
	)
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenVaults := make(map[string]bool)

	for _, vault := range gs.Vaults {
		if seenVaults[vault.Key] {
			return fmt.Errorf("duplicate vault for name %s", vault.Key)
		}

		seenVaults[vault.Key] = true
	}

	for _, lock := range gs.Locks {
		if !seenVaults[lock.Key] {
			return fmt.Errorf("no vault %s for the lock", lock.Key)
		}
	}

	for _, stake := range gs.Stakes {
		if _, err := sdk.AccAddressFromBech32(stake.StakerAddress); err != nil {
			return err
		}

		if err := stake.Coins.Validate(); err != nil {
			return err
		}
	}

	for _, stake := range gs.Stakes {
		if _, err := sdk.AccAddressFromBech32(stake.StakerAddress); err != nil {
			return err
		}

		if err := stake.Coins.Validate(); err != nil {
			return err
		}
	}

	return nil
}
