package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
)

var _ types.UnpackInterfacesMessage = Packet{}

func NewPacket(
	tunnelID uint64,
	sequence uint64,
	signalPrices []SignalPrice,
	createdAt int64,
) Packet {
	return Packet{
		TunnelID:      tunnelID,
		Sequence:      sequence,
		SignalPrices:  signalPrices,
		PacketContent: nil,
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

func (p Packet) EncodeTss(
	destinationChainID string,
	destinationContractAddress string,
	encoder Encoder,
) ([]byte, error) {
	encodingPacket, err := NewTssEncodingPacket(
		p,
		destinationChainID,
		destinationContractAddress,
		encoder,
	)
	if err != nil {
		return nil, err
	}

	switch encoder {
	case ENCODER_FIXED_POINT_ABI, ENCODER_TICK_ABI:
		return encodingPacket.EncodeABI()
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder mode: %s", encoder.String())
	}
}
