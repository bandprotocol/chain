package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &RouterRoute{}

// NewRouterRoute creates a new RouterRoute instance.
func NewRouterRoute(
	fund sdk.Coin,
	bridgeContractAddress string,
	destChinID string,
	destContractAddress string,
	destGasLimit uint64,
	destGasPrice uint64,
) *RouterRoute {
	return &RouterRoute{
		Fund:                  fund,
		BridgeContractAddress: bridgeContractAddress,
		DestChainID:           destChinID,
		DestContractAddress:   destContractAddress,
		DestGasLimit:          destGasLimit,
		DestGasPrice:          destGasPrice,
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
