package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	packetABI, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
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

	packetArgs = abi.Arguments{
		{Type: packetABI},
	}
)

// PacketABI represents the Packet that will be used for encoding a tunnel packet into abi format.
type PacketABI struct {
	TunnelID     uint64
	Sequence     uint64
	SignalPrices []feedstypes.RelayPrice
	CreatedAt    int64
}

// NewPacketABI returns a new PacketABI object
func NewPacketABI(
	tunnelID uint64,
	sequence uint64,
	signalPrices []feedstypes.RelayPrice,
	createdAt int64,
) PacketABI {
	return PacketABI{
		TunnelID:     tunnelID,
		Sequence:     sequence,
		SignalPrices: signalPrices,
		CreatedAt:    createdAt,
	}
}

// EncodingPacketABI encodes the packet to abi message
func EncodingPacketABI(p Packet) ([]byte, error) {
	var signalPrices []feedstypes.RelayPrice
	for _, sp := range p.Prices {
		signalPrices = append(signalPrices, feedstypes.RelayPrice{
			SignalID: stringToBytes32(sp.SignalID),
			Price:    sp.Price,
		})
	}

	abiPacket := NewPacketABI(p.TunnelID, p.Sequence, signalPrices, p.CreatedAt)
	return packetArgs.Pack(&abiPacket)
}

// stringToBytes32 converts a string to a fixed size byte array. If the string is longer than
// 32 bytes, it will be truncated to the first 32 bytes. If the string is shorter than 32 bytes,
// it will be padded with 0s at the beginning.
func stringToBytes32(str string) [32]byte {
	maxLen := len(str)
	if maxLen > 32 {
		maxLen = 32
	}

	var byteArray [32]byte
	copy(byteArray[32-maxLen:], str[:maxLen])
	return byteArray
}
