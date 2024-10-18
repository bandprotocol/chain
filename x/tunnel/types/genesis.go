package types

import (
	"fmt"

	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(
	params Params,
	portID string,
	tunnelCount uint64,
	tunnels []Tunnel,
	latestSignalPricesList []LatestSignalPrices,
	totalFees TotalFees,
) *GenesisState {
	return &GenesisState{
		Params:                 params,
		PortID:                 portID,
		TunnelCount:            tunnelCount,
		Tunnels:                tunnels,
		LatestSignalPricesList: latestSignalPricesList,
		TotalFees:              totalFees,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), PortID, 0, []Tunnel{}, []LatestSignalPrices{}, TotalFees{})
}

// ValidateGenesis validates the provided genesis state.
func ValidateGenesis(data GenesisState) error {
	// validate the port ID
	if err := host.PortIdentifierValidator(data.PortID); err != nil {
		return ErrInvalidGenesis.Wrapf("invalid port ID: %s", err.Error())
	}

	// validate the tunnel count
	if uint64(len(data.Tunnels)) != data.TunnelCount {
		return ErrInvalidGenesis.Wrapf("length of tunnels does not match tunnel count")
	}

	// validate the tunnel IDs
	for _, t := range data.Tunnels {
		if t.ID > data.TunnelCount {
			return ErrInvalidGenesis.Wrapf("tunnel count mismatch in tunnels")
		}
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
	lsps []LatestSignalPrices,
) error {
	if len(tunnels) != len(lsps) {
		return fmt.Errorf("tunnels and latest signal prices list length mismatch")
	}

	tunnelMap := make(map[uint64]bool)
	for _, t := range tunnels {
		tunnelMap[t.ID] = true
	}

	for _, lsp := range lsps {
		if _, ok := tunnelMap[lsp.TunnelID]; !ok {
			return fmt.Errorf("tunnel ID %d not found in tunnels", lsp.TunnelID)
		}
		if err := lsp.Validate(); err != nil {
			return err
		}
	}
	return nil
}
