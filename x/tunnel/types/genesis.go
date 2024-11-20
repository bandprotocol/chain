package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(
	params Params,
	tunnelCount uint64,
	tunnels []Tunnel,
	totalFees TotalFees,
) *GenesisState {
	return &GenesisState{
		Params:      params,
		TunnelCount: tunnelCount,
		Tunnels:     tunnels,
		TotalFees:   totalFees,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, []Tunnel{}, TotalFees{})
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

// Total returns the total fees
func (tf TotalFees) Total() sdk.Coins {
	return tf.TotalPacketFee
}
