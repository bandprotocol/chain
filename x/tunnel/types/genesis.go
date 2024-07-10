package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(
	params Params,
	tunnelCount uint64,
	tssPacketCount uint64,
	axelarPacketCount uint64,
	tunnels []Tunnel,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		TunnelCount:       tunnelCount,
		TssPacketCount:    tssPacketCount,
		AxelarPacketCount: axelarPacketCount,
		Tunnels:           tunnels,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, 0, 0, []Tunnel{})
}
