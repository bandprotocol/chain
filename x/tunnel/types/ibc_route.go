package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// IBCRoute defines the IBC route for the tunnel module
var _ Route = &IBCRoute{}

// Route defines the IBC route for the tunnel module
func (r *IBCRoute) ValidateBasic() error {
	return nil
}

// NewIBCPacket creates a new IBCPacket instance
func NewIBCPacket(
	tunnelID uint64,
	nonce uint64,
	feedType types.FeedType,
	signalPriceInfos []SignalPriceInfo,
	channelID string,
	createdAt int64,
) IBCPacket {
	return IBCPacket{
		TunnelID:         tunnelID,
		Nonce:            nonce,
		FeedType:         feedType,
		SignalPriceInfos: signalPriceInfos,
		ChannelID:        channelID,
		CreatedAt:        createdAt,
	}
}

// GetBytes returns the raw bytes of the IBCPacket
func (p IBCPacket) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
