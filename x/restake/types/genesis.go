package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(vaults []Vault, locks []Lock) *GenesisState {
	return &GenesisState{
		Vaults: vaults,
		Locks:  locks,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Vaults: []Vault{},
		Locks:  []Lock{},
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	seenVaults := make(map[string]bool)
	totalPowers := make(map[string]sdkmath.Int)

	for _, lock := range gs.Locks {
		total, ok := totalPowers[lock.Key]
		if !ok {
			total = sdkmath.NewInt(0)
		}

		totalPowers[lock.Key] = total.Add(lock.Power)
	}

	for _, vault := range gs.Vaults {
		if seenVaults[vault.Key] {
			return fmt.Errorf("duplicate vault for name %s", vault.Key)
		}

		seenVaults[vault.Key] = true

		// if vault is active, total power must be equal.
		if vault.IsActive && !vault.TotalPower.Equal(totalPowers[vault.Key]) {
			return fmt.Errorf(
				"genesis total_power is incorrect, expected %v, got %v",
				vault.TotalPower,
				totalPowers[vault.Key],
			)
		}
	}

	return nil
}
