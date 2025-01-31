package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	hyperlaneStridePacketABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "TunnelID", Type: "uint64"},
		{Name: "Sequence", Type: "uint64"},
		{
			Name: "SignalPrices",
			Type: "tuple[]",
			Components: []abi.ArgumentMarshaling{
				{Name: "SignalID", Type: "bytes32"},
				{Name: "Price", Type: "uint64"},
			},
		},
		{Name: "CreatedAt", Type: "int64"},
	})

	hyperlaneStrideRelayPacketABI = abi.NewMethod(
		"relayPacket",
		"relayPacket",
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{Name: "packet", Type: hyperlaneStridePacketABI},
		},
		nil,
	)
)

// EncoderHyperlaneStridePacket represents the Packet that will be used for encoding a message.
type EncoderHyperlaneStridePacket struct {
	TunnelID     uint64
	Sequence     uint64
	SignalPrices []feedstypes.RelayPrice
	CreatedAt    int64
}

// NewEncoderHyperlaneStridePacket returns a new EncoderHyperlaneStridePacket object
func NewEncoderHyperlaneStridePacket(
	tunnelID uint64,
	sequence uint64,
	signalPrices []feedstypes.RelayPrice,
	createdAt int64,
) EncoderHyperlaneStridePacket {
	return EncoderHyperlaneStridePacket{
		TunnelID:     tunnelID,
		Sequence:     sequence,
		SignalPrices: signalPrices,
		CreatedAt:    createdAt,
	}
}

// EncodingHyperlaneStride encodes the packet to hyperlane axelar message
func EncodingHyperlaneStride(p Packet) ([]byte, error) {
	var signalPrices []feedstypes.RelayPrice
	for _, sp := range p.Prices {
		signalPrices = append(signalPrices, feedstypes.RelayPrice{
			SignalID: stringToBytes32(sp.SignalID),
			Price:    sp.Price,
		})
	}

	hp := NewEncoderHyperlaneStridePacket(p.TunnelID, p.Sequence, signalPrices, p.CreatedAt)

	packetBytes, err := hyperlaneStrideRelayPacketABI.Inputs.Pack(&hp)
	if err != nil {
		return nil, err
	}

	// prepend the method ID (function selector)
	return append(hyperlaneStrideRelayPacketABI.ID, packetBytes...), nil
}
