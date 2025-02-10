package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	tssPacketABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "Sequence", Type: "uint64"},
		{
			Name:         "RelayPrices",
			Type:         "tuple[]",
			InternalType: "struct Prices[]",
			Components: []abi.ArgumentMarshaling{
				{Name: "SignalID", Type: "bytes32"},
				{Name: "Price", Type: "uint64"},
			},
		},
		{Name: "CreatedAt", Type: "int64"},
	})

	tssPacketArgs = abi.Arguments{
		{Type: tssPacketABI, Name: "packet"},
	}
)

// TSSPacket represents the Packet that will be used for encoding a tss message.
type TSSPacket struct {
	Sequence    uint64
	RelayPrices []feedstypes.RelayPrice
	CreatedAt   int64
}

// NewTSSPacket returns a new TssPacket object
func NewTSSPacket(
	sequence uint64,
	relayPrices []feedstypes.RelayPrice,
	createdAt int64,
) TSSPacket {
	return TSSPacket{
		Sequence:    sequence,
		RelayPrices: relayPrices,
		CreatedAt:   createdAt,
	}
}

// EncodeTSS encodes the packet to tss message
func EncodeTSS(
	sequence uint64,
	prices []feedstypes.Price,
	createdAt int64,
	encoder feedstypes.Encoder,
) ([]byte, error) {
	switch encoder {
	case feedstypes.ENCODER_FIXED_POINT_ABI:
		relayPrices, err := feedstypes.ToRelayPrices(prices)
		if err != nil {
			return nil, err
		}

		tssPacket := NewTSSPacket(sequence, relayPrices, createdAt)

		bz, err := tssPacketArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderFixedPointABIPrefix), bz...), nil
	case feedstypes.ENCODER_TICK_ABI:
		relayPrices, err := feedstypes.ToRelayTickPrices(prices)
		if err != nil {
			return nil, err
		}

		tssPacket := NewTSSPacket(sequence, relayPrices, createdAt)

		bz, err := tssPacketArgs.Pack(&tssPacket)
		if err != nil {
			return nil, err
		}

		return append([]byte(feedstypes.EncoderTickABIPrefix), bz...), nil
	default:
		return nil, ErrInvalidEncoder.Wrapf("invalid encoder mode: %s", encoder.String())
	}
}
