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
			Name:         "TssPrices",
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

// TssPacket represents the Packet that will be used for encoding a tss message.
type TssPacket struct {
	TunnelID                   uint64
	Sequence                   uint64
	DestinationChainID         string
	DestinationContractAddress string
	TssPrices                  []feedstypes.TssPrice
	CreatedAt                  int64
}

// NewTssPacket returns a new TssPacket object
func NewTssPacket(
	tunnelID uint64,
	sequence uint64,
	destinationChainID string,
	destinationContractAddress string,
	tssPrices []feedstypes.TssPrice,
	createdAt int64,
) TssPacket {
	return TssPacket{
		TunnelID:                   tunnelID,
		Sequence:                   sequence,
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		TssPrices:                  tssPrices,
		CreatedAt:                  createdAt,
	}
}

// EncodeTss encodes the packet to tss message
func EncodeTss(
	packet Packet,
	destinationChainID string,
	destinationContractAddress string,
	encoder feedstypes.Encoder,
) ([]byte, error) {
	switch encoder {
	case feedstypes.ENCODER_FIXED_POINT_ABI:
		tssPrices, err := feedstypes.ToTssPrices(packet.Prices)
		if err != nil {
			return nil, err
		}

		tssPacket := NewTssPacket(
			packet.TunnelID,
			packet.Sequence,
			destinationChainID,
			destinationContractAddress,
			tssPrices,
			packet.CreatedAt,
		)

		bz, err := packetArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderFixedPointABIPrefix), bz...), nil
	case feedstypes.ENCODER_TICK_ABI:
		tssPrices, err := feedstypes.ToTssTickPrices(packet.Prices)
		if err != nil {
			return nil, err
		}
		tssPacket := NewTssPacket(
			packet.TunnelID,
			packet.Sequence,
			destinationChainID,
			destinationContractAddress,
			tssPrices,
			packet.CreatedAt,
		)

		bz, err := packetArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderTickABIPrefix), bz...), nil
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder mode: %s", encoder.String())
	}
}
