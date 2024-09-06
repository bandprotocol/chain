package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
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
func NewTssPacket(p Packet) *TssPacket {
	var tssSignalPrices []TssSignalPrice
	for _, sp := range p.SignalPrices {
		var signalID [32]byte
		copy(signalID[:], sp.SignalID)
		tssSignalPrices = append(tssSignalPrices, TssSignalPrice{
			SignalID: signalID,
			Price:    sp.Price,
		})
	}

	return &TssPacket{
		TunnelID:     p.TunnelID,
		Nonce:        p.Nonce,
		SignalPrices: tssSignalPrices,
		CreatedAt:    p.CreatedAt,
	}
}

// EncodeAbi encodes the TssPacket into bytes
func (p TssPacket) EncodeAbi() ([]byte, error) {
	return packetArgs.Pack(&p)
}
