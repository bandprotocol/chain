package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

var _ types.UnpackInterfacesMessage = Packet{}

func NewPacket(
	tunnelID uint64,
	nonce uint64,
	signalPriceInfos []SignalPriceInfo,
	createdAt int64,
) Packet {
	return Packet{
		TunnelID:         tunnelID,
		Nonce:            nonce,
		SignalPriceInfos: signalPriceInfos,
		CreatedAt:        createdAt,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Packet) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var packetContent PacketContentI
	return unpacker.UnpackAny(p.PacketContent, &packetContent)
}

// SetPacketContent sets the packet content of the packet.
func (p *Packet) SetPacketContent(packetContent PacketContentI) error {
	msg, ok := packetContent.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	p.PacketContent = any

	return nil
}

// GetGrant unpacks allowance
func (p Packet) GetContent() (PacketContentI, error) {
	packetContent, ok := p.PacketContent.GetCachedValue().(PacketContentI)
	if !ok {
		return nil, ErrNoPacketContent.Wrapf("tunnelID: %d, nonce: %d", p.TunnelID, p.Nonce)
	}

	return packetContent, nil
}
