package types

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// hookCoinDenom defines the coin denomination to be used for minting and transferring coins in the IBC hook route.
	hookCoinDenom = "uhook"
	// TransferAmount defines the amount of coins to be transferred in the IBC hook route.
	TransferAmount = sdk.NewInt64Coin(hookCoinDenom, 1)
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
	return nil
}

// NewIBCHookPacketReceipt creates a new IBCHookPacketReceipt instance. It is used to create the IBC hook packet receipt data
func NewIBCHookPacketReceipt(sequence uint64) *IBCHookPacketReceipt {
	return &IBCHookPacketReceipt{
		Sequence: sequence,
	}
}
