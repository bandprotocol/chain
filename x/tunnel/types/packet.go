package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.UnpackInterfacesMessage = Packet{}

func NewPacket(
	tunnelID uint64,
	sequence uint64,
	signalPrices []SignalPrice,
	baseFee sdk.Coins,
	routeFee sdk.Coins,
	createdAt int64,
) Packet {
	return Packet{
		TunnelID:      tunnelID,
		Sequence:      sequence,
		SignalPrices:  signalPrices,
		PacketContent: nil,
		BaseFee:       baseFee,
		RouteFee:      routeFee,
		CreatedAt:     createdAt,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Packet) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var packetContent PacketContentI
	return unpacker.UnpackAny(p.PacketContent, &packetContent)
}

// SetPacketContent sets the packet content of the packet.
func (p *Packet) SetPacketContent(packetContent PacketContentI) error {
	any, err := types.NewAnyWithValue(packetContent)
	if err != nil {
		return err
	}
	p.PacketContent = any

	return nil
}

// GetContent returns the content of the packet.
func (p Packet) GetContent() (PacketContentI, error) {
	packetContent, ok := p.PacketContent.GetCachedValue().(PacketContentI)
	if !ok {
		return nil, ErrNoPacketContent.Wrapf("tunnelID: %d, sequence: %d", p.TunnelID, p.Sequence)
	}

	return packetContent, nil
}
