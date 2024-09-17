package types

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(
	params Params,
	portID string,
	tunnelCount uint64,
	tunnels []Tunnel,
	signalPricesInfos []SignalPricesInfo,
	totalFees TotalFees,
) *GenesisState {
	return &GenesisState{
		Params:            params,
		PortID:            portID,
		TunnelCount:       tunnelCount,
		Tunnels:           tunnels,
		SignalPricesInfos: signalPricesInfos,
		TotalFees:         totalFees,
	}
}

// DefaultGenesisState gets the raw genesis raw message for testing
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), PortID, 0, []Tunnel{}, []SignalPricesInfo{}, TotalFees{})
}
