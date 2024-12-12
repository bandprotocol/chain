package types

import (
	fmt "fmt"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &RouterRoute{}

// NewRouterRoute creates a new RouterRoute instance.
func NewRouterRoute(
	channelID string,
	fund sdk.Coin,
	bridgeContractAddress string,
	destChinID string,
	destContractAddress string,
	destGasLimit uint64,
	destGasPrice uint64,
) *RouterRoute {
	return &RouterRoute{
		ChannelID:             channelID,
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
	// Validate the ChannelID format
	if r.ChannelID != "" && !channeltypes.IsChannelIDFormat(r.ChannelID) {
		return fmt.Errorf("channel identifier is not in the format: `channel-{N}` or be empty string")
	}
	return nil
}

// NewRouterPacketReceipt creates a new RouterPacketReceipt instance.
func NewRouterPacketReceipt(sequence uint64) *RouterPacketReceipt {
	return &RouterPacketReceipt{
		Sequence: sequence,
	}
}
