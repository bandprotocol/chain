package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	packetABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "Sequence", Type: "uint64"},
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
	Sequence  uint64
	TssPrices []feedstypes.TssPrice
	CreatedAt int64
}

// NewTssPacket returns a new TssPacket object
func NewTssPacket(
	sequence uint64,
	tssPrices []feedstypes.TssPrice,
	createdAt int64,
) TssPacket {
	return TssPacket{
		Sequence:  sequence,
		TssPrices: tssPrices,
		CreatedAt: createdAt,
	}
}

// EncodeTss encodes the packet to tss message
func EncodeTss(
	sequence uint64,
	prices []feedstypes.Price,
	createdAt int64,
	encoder feedstypes.Encoder,
) ([]byte, error) {
	switch encoder {
	case feedstypes.ENCODER_FIXED_POINT_ABI:
		tssPrices, err := feedstypes.ToTssPrices(prices)
		if err != nil {
			return nil, err
		}

		tssPacket := NewTssPacket(sequence, tssPrices, createdAt)

		bz, err := packetArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderFixedPointABIPrefix), bz...), nil
	case feedstypes.ENCODER_TICK_ABI:
		tssPrices, err := feedstypes.ToTssTickPrices(prices)
		if err != nil {
			return nil, err
		}

		tssPacket := NewTssPacket(sequence, tssPrices, createdAt)

		bz, err := packetArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderTickABIPrefix), bz...), nil
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder mode: %s", encoder.String())
	}
}
