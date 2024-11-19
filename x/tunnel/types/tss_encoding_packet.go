package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	packetABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "TunnelID", Type: "uint64"},
		{Name: "Sequence", Type: "uint64"},
		{Name: "DestinationChainID", Type: "string"},
		{Name: "DestinationContractAddress", Type: "string"},
		{
			Name:         "PriceEncoders",
			Type:         "tuple[]",
			InternalType: "struct Prices[]",
			Components: []abi.ArgumentMarshaling{
				{Name: "SignalID", Type: "bytes32"},
				{Name: "Price", Type: "uint64"},
			},
		},
		{Name: "CreatedAt", Type: "int64"},
	})

	packetArgs = abi.Arguments{
		{Type: packetABI, Name: "packet"},
	}
)

// TssEncodingPacket represents the Packet that will be used for encoding a message.
type TssEncodingPacket struct {
	TunnelID                   uint64
	Sequence                   uint64
	DestinationChainID         string
	DestinationContractAddress string
	PriceEncoders              feedstypes.PriceEncoders
	CreatedAt                  int64
}

// NewTssEncodingPacket returns a new TssEncodingPacket object
func NewTssEncodingPacket(
	packet Packet,
	destinationChainID string,
	destinationContractAddress string,
	encoder feedstypes.Encoder,
) (*TssEncodingPacket, error) {
	priceEncoders, err := feedstypes.ToPriceEncoders(packet.Prices, encoder)
	if err != nil {
		return nil, err
	}

	return &TssEncodingPacket{
		TunnelID:                   packet.TunnelID,
		Sequence:                   packet.Sequence,
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		PriceEncoders:              priceEncoders,
		CreatedAt:                  packet.CreatedAt,
	}, nil
}

// EncodeABI encodes the encoding packet into bytes via ABI encoding
func (p TssEncodingPacket) EncodeABI() ([]byte, error) {
	return packetArgs.Pack(&p)
}
