package types

import (
	fmt "fmt"

	sdkmath "cosmossdk.io/math"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(keys []Key, locks []Lock) *GenesisState {
	return &GenesisState{
		Keys:  keys,
		Locks: locks,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Keys:  []Key{},
		Locks: []Lock{},
	}
}

// Validate performs basic validation of genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	seenKeys := make(map[string]bool)
	totalPowers := make(map[string]sdkmath.Int)

	for _, lock := range gs.Locks {
		_, ok := totalPowers[lock.Key]
		if !ok {
			totalPowers[lock.Key] = sdkmath.NewInt(0)
		}

		totalPowers[lock.Key] = totalPowers[lock.Key].Add(lock.Power)
	}

	for _, key := range gs.Keys {
		if seenKeys[key.Name] {
			return fmt.Errorf("duplicate key for name %s", key.Name)
		}

		seenKeys[key.Name] = true

		// if key is active, total power must be equal.
		if key.IsActive && !key.TotalPower.Equal(totalPowers[key.Name]) {
			return fmt.Errorf(
				"genesis total_power is incorrect, expected %v, got %v",
				key.TotalPower,
				totalPowers[key.Name],
			)
		}
	}

	return nil
}
