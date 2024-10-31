package types

import (
	"fmt"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(
	params Params,
	tunnelCount uint64,
	tunnels []Tunnel,
	latestSignalPricesList []LatestSignalPrices,
	totalFees TotalFees,
) *GenesisState {
	return &GenesisState{
		Params:                 params,
		TunnelCount:            tunnelCount,
		Tunnels:                tunnels,
		LatestSignalPricesList: latestSignalPricesList,
		TotalFees:              totalFees,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, []Tunnel{}, []LatestSignalPrices{}, TotalFees{})
}

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data GenesisState) error {
	// validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return ErrInvalidGenesis.Wrapf("length of tunnels does not match tunnel count")
	}

	// validate the tunnel IDs
	tunnelIDMap := make(map[uint64]bool)
	for _, t := range data.Tunnels {
		if t.ID > data.TunnelCount {
			return ErrInvalidGenesis.Wrapf("tunnel count mismatch in tunnels")
		}
		if _, exists := tunnelIDMap[t.ID]; exists {
			return ErrInvalidGenesis.Wrapf("duplicate tunnel ID found: %d", t.ID)
		}
		tunnelIDMap[t.ID] = true
	}

	// validate the latest signal prices count
	if len(data.LatestSignalPricesList) != int(data.TunnelCount) {
		return ErrInvalidGenesis.Wrapf("tunnel count mismatch in latest signal prices")
	}

	// validate latest signal prices
	if err := validateLastSignalPricesList(data.Tunnels, data.LatestSignalPricesList); err != nil {
		return ErrInvalidGenesis.Wrapf("invalid latest signal prices: %s", err.Error())
	}

	// validate the total fees
	if err := data.TotalFees.Validate(); err != nil {
		return ErrInvalidGenesis.Wrapf("invalid total fees: %s", err.Error())
	}

	return data.Params.Validate()
}

// Validate validates the total fees
func (tf TotalFees) Validate() error {
	if !tf.TotalPacketFee.IsValid() {
		return fmt.Errorf("invalid total packet fee: %s", tf.TotalPacketFee)
	}
	return nil
}

// validateLastSignalPricesList validates the latest signal prices list.
func validateLastSignalPricesList(
	tunnels []Tunnel,
	latestSignalPricesList []LatestSignalPrices,
) error {
	if len(tunnels) != len(latestSignalPricesList) {
		return fmt.Errorf("tunnels and latest signal prices list length mismatch")
	}

	tunnelIDMap := make(map[uint64]bool)
	for _, t := range tunnels {
		tunnelIDMap[t.ID] = true
	}

	for _, latestSignalPrices := range latestSignalPricesList {
		if !tunnelIDMap[latestSignalPrices.TunnelID] {
			return fmt.Errorf("tunnel ID %d not found in tunnels", latestSignalPrices.TunnelID)
		}
		if err := latestSignalPrices.Validate(); err != nil {
			return err
		}
	}

	return nil
}
