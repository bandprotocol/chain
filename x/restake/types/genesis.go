package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
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
