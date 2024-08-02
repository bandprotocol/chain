package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(
	portID string,
	params Params,
	tunnelCount uint64,
	tunnels []Tunnel,
) *GenesisState {
	return &GenesisState{
		PortID:      portID,
		Params:      params,
		TunnelCount: tunnelCount,
		Tunnels:     tunnels,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(PortID, DefaultParams(), 0, []Tunnel{})
}
