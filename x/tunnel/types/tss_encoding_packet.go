package types

import (
	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/bandprotocol/chain/v3/pkg/tickmath"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	packetABI, _ = abi.NewType("tuple", "result", []abi.ArgumentMarshaling{
		{Name: "TunnelID", Type: "uint64"},
		{Name: "Sequence", Type: "uint64"},
		{Name: "DestinationChainID", Type: "string"},
		{Name: "DestinationContractAddress", Type: "string"},
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

// TssEncodingPrice represents the price that will be used for encoding a message.
type TssEncodingPrice struct {
	SignalID [32]byte
	Price    uint64
}

// TssEncodingPacket represents the Packet that will be used for encoding a message.
type TssEncodingPacket struct {
	TunnelID                   uint64
	Sequence                   uint64
	DestinationChainID         string
	DestinationContractAddress string
	SignalPrices               []TssEncodingPrice
	CreatedAt                  int64
}

// NewTssEncodingPacket returns a new TssEncodingPacket object
func NewTssEncodingPacket(
	p Packet,
	destinationChainID string,
	destinationContractAddress string,
	encoder feedstypes.Encoder,
) (*TssEncodingPacket, error) {
	var signalPrices []TssEncodingPrice
	for _, sp := range p.Prices {
		price := sp.Price
		if encoder == feedstypes.ENCODER_TICK_ABI && price != 0 {
			tick, err := tickmath.PriceToTick(price)
			if err != nil {
				return nil, err
			}
			price = tick
		}

		signalPrices = append(signalPrices, TssEncodingPrice{
			SignalID: stringToBytes32(sp.SignalID),
			Price:    price,
		})
	}

	return &TssEncodingPacket{
		TunnelID:                   p.TunnelID,
		Sequence:                   p.Sequence,
		DestinationChainID:         destinationChainID,
		DestinationContractAddress: destinationContractAddress,
		SignalPrices:               signalPrices,
		CreatedAt:                  p.CreatedAt,
	}, nil
}

// EncodeABI encodes the encoding packet into bytes via ABI encoding
func (p TssEncodingPacket) EncodeABI() ([]byte, error) {
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
