package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	tunnelIDs := make(map[uint64]bool)
	for _, t := range data.Tunnels {
		if t.ID > data.TunnelCount {
			return ErrInvalidGenesis.Wrapf("tunnel count mismatch in tunnels")
		}
		if _, exists := tunnelIDs[t.ID]; exists {
			return ErrInvalidGenesis.Wrapf("duplicate tunnel ID found: %d", t.ID)
		}
		tunnelIDs[t.ID] = true
	}

	// validate the latest signal prices count
	if len(data.LatestSignalPricesList) != int(data.TunnelCount) {
		return ErrInvalidGenesis.Wrapf("tunnel count mismatch in latest signal prices")
	}

	// validate latest signal prices
	if err := validateLastestSignalPricesList(data.Tunnels, data.LatestSignalPricesList); err != nil {
		return ErrInvalidGenesis.Wrapf("invalid latest signal prices: %s", err.Error())
	}

	// validate no duplicated deposits
	type depositKey struct {
		TunnelID  uint64
		Depositor string
	}
	deposits := make(map[depositKey]bool)
	tunnelDeposit := make(map[uint64]sdk.Coins)
	for _, d := range data.Deposits {
		if _, ok := tunnelIDs[d.TunnelID]; !ok {
			return ErrInvalidGenesis.Wrapf("deposit has non-existent tunnel id: %d, deposit: %+v", d.TunnelID, d)
		}

		dk := depositKey{d.TunnelID, d.Depositor}
		if _, ok := deposits[dk]; ok {
			return ErrInvalidGenesis.Wrapf("duplicate deposit: %v", d)
		}

		deposits[dk] = true
		tunnelDeposit[d.TunnelID] = tunnelDeposit[d.TunnelID].Add(d.Amount...)
	}

	// validate total deposit on tunnels with deposits
	for _, t := range data.Tunnels {
		if !t.TotalDeposit.Equal(tunnelDeposit[t.ID]) {
			return ErrInvalidGenesis.Wrapf("deposits mismatch total deposit for tunnel %d", t.ID)
		}
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

// validateLastestSignalPricesList validates the latest signal prices list.
func validateLastestSignalPricesList(
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
