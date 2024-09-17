package types

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
