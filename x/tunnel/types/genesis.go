package types

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(
	params Params,
	tunnelCount uint64,
	tunnels []Tunnel,
	activeTunnelIDs []uint64,
	signalPricesInfos []SignalPricesInfo,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		TunnelCount:       tunnelCount,
		Tunnels:           tunnels,
		ActiveTunnelIDs:   activeTunnelIDs,
		SignalPricesInfos: signalPricesInfos,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), 0, []Tunnel{}, []uint64{}, []SignalPricesInfo{})
}
