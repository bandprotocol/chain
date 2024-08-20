package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBCPacket defines the packet sent over the IBC channel
func NewIBCPacketResult(
	tunnelID uint64,
	nonce uint64,
	signalPriceInfos []SignalPriceInfo,
) IBCPacketResult {
	return IBCPacketResult{
		TunnelID:         tunnelID,
		Nonce:            nonce,
		SignalPriceInfos: signalPriceInfos,
	}
}

// GetBytes returns the IBCPacketResult bytes
func (p IBCPacketResult) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&p))
}
