package types

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

// IBCRoute defines the IBC route for the tunnel module
var _ RouteI = &IBCRoute{}

// NewIBCRoute creates a new IBCRoute instance.
func NewIBCRoute(channelID string) *IBCRoute {
	return &IBCRoute{
		ChannelID: channelID,
	}
}

// Route defines the IBC route for the tunnel module
func (r *IBCRoute) ValidateBasic() error {
	// Validate the ChannelID format
	if !channeltypes.IsChannelIDFormat(r.ChannelID) {
		return fmt.Errorf("channel identifier is not in the format: `channel-{N}`")
	}
	return nil
}

// NewIBCPacketReceipt creates a new IBCPacketReceipt instance.
func NewIBCPacketReceipt(sequence uint64) *IBCPacketReceipt {
	return &IBCPacketReceipt{
		Sequence: sequence,
	}
}

// NewTunnelPricesPacketData creates a new TunnelPricesPacketData instance.
func NewTunnelPricesPacketData(
	tunnelID uint64,
	sequence uint64,
	prices []feedstypes.Price,
	created_at int64,
) TunnelPricesPacketData {
	return TunnelPricesPacketData{
		TunnelID:  tunnelID,
		Sequence:  sequence,
		Prices:    prices,
		CreatedAt: created_at,
	}
}

// GetBytes is a helper for serialising
func (p TunnelPricesPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
