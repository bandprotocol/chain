package types

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &RouterRoute{}

// NewRouterRoute creates a new RouterRoute instance.
func NewRouterRoute(
	destinationChinID string,
	destinationContractAddress string,
	destinationGasLimit uint64,
) *RouterRoute {
	return &RouterRoute{
		DestinationChainID:         destinationChinID,
		DestinationContractAddress: destinationContractAddress,
		DestinationGasLimit:        destinationGasLimit,
	}
}

// ValidateBasic validates the RouterRoute
func (r *RouterRoute) ValidateBasic() error {
	return nil
}

// NewRouterPacketReceipt creates a new RouterPacketReceipt instance.
func NewRouterPacketReceipt(sequence uint64) *RouterPacketReceipt {
	return &RouterPacketReceipt{
		Sequence: sequence,
	}
}
