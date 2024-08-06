package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func NewIBCPacket(
	tunnelID uint64,
	nonce uint64,
	feedType types.FeedType,
	signalPriceInfos []SignalPriceInfo,
	portID string,
	channelID string,
	createdAt uint64,
) IBCPacket {
	return IBCPacket{
		TunnelID:         tunnelID,
		Nonce:            nonce,
		FeedType:         feedType,
		SignalPriceInfos: signalPriceInfos,
		PortID:           portID,
		ChannelID:        channelID,
		CreatedAt:        createdAt,
	}
}

func (p IBCPacket) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
