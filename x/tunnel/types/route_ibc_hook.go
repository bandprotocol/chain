package types

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

var (
	// HookDenomPrefix defines the prefix for the IBC hook denom
	HookDenomPrefix = "tunnel-"
	// HookTransferAmount defines the amount to transfer for the IBC hook
	HookTransferAmount = int64(1)
)

// IBCRoute defines the IBC route for the tunnel module
var _ RouteI = &IBCHookRoute{}

// NewIBCHookRoute creates a new IBCHookRoute instance. It is used to create the IBC hook route data
func NewIBCHookRoute(channelID, destinationContractAddress string) *IBCHookRoute {
	return &IBCHookRoute{
		ChannelID:                  channelID,
		DestinationContractAddress: destinationContractAddress,
	}
}

// Route defines the IBC route for the tunnel module
func (r *IBCHookRoute) ValidateBasic() error {
	// Validate the ChannelID format
	if !channeltypes.IsChannelIDFormat(r.ChannelID) {
		return fmt.Errorf("channel identifier is not in the format: `channel-{N}`")
	}

	// Validate the DestinationContractAddress cannot be empty
	if r.DestinationContractAddress == "" {
		return fmt.Errorf("destination contract address cannot be empty")
	}

	return nil
}

// NewIBCHookPacketReceipt creates a new IBCHookPacketReceipt instance. It is used to create the IBC hook packet receipt data
func NewIBCHookPacketReceipt(sequence uint64) *IBCHookPacketReceipt {
	return &IBCHookPacketReceipt{
		Sequence: sequence,
	}
}

// FormatHookDenomIdentifier returns the hook denom identifier based on the tunnel ID
func FormatHookDenomIdentifier(tunnelID uint64) string {
	return fmt.Sprintf("%s%d", HookDenomPrefix, tunnelID)
}
