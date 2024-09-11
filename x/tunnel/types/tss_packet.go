package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/bandprotocol/chain/v2/pkg/tickmath"
)

var (
	packetABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "TunnelID", Type: "uint64"},
		{Name: "Nonce", Type: "uint64"},
		{
			Name:         "SignalPrices",
			Type:         "tuple[]",
			InternalType: "struct SignalPrices[]",
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

type TssSignalPrice struct {
	SignalID [32]byte
	Price    uint64
}

type TssPacket struct {
	TunnelID     uint64
	Nonce        uint64
	SignalPrices []TssSignalPrice
	CreatedAt    int64
}

// NewTssPacket returns a new TssPacket object
func NewTssPacket(p Packet, encoder Encoder) (*TssPacket, error) {
	var tssSignalPrices []TssSignalPrice
	for _, sp := range p.SignalPrices {
		price := sp.Price
		if encoder == ENCODER_TICK_ABI && price != 0 {
			tick, err := tickmath.PriceToTick(price)
			if err != nil {
				return nil, err
			}
			price = tick
		}

		tssSignalPrices = append(tssSignalPrices, TssSignalPrice{
			SignalID: stringToBytes32(sp.SignalID),
			Price:    price,
		})
	}

	return &TssPacket{
		TunnelID:     p.TunnelID,
		Nonce:        p.Nonce,
		SignalPrices: tssSignalPrices,
		CreatedAt:    p.CreatedAt,
	}, nil
}

// EncodeAbi encodes the TssPacket into bytes
func (p TssPacket) EncodeAbi() ([]byte, error) {
	return packetArgs.Pack(&p)
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
