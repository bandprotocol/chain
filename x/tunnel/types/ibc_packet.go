package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
)

// IBCPacket defines the packet sent over the IBC channel
func NewIBCPacket(
	tunnelID uint64,
	nonce uint64,
	feedType feedstypes.FeedType,
	signalPriceInfos []SignalPriceInfo,
	ibcPacketContent IBCPacketContent,
	createdAt int64,
) IBCPacket {
	return IBCPacket{
		TunnelID:         tunnelID,
		Nonce:            nonce,
		FeedType:         feedType,
		SignalPriceInfos: signalPriceInfos,
		IBCPacketContent: ibcPacketContent,
		CreatedAt:        createdAt,
	}
}

// GetBytes returns the raw bytes of the packet
func (p IBCPacket) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
