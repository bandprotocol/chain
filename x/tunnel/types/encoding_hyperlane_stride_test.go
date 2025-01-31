package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestEncodeABI(t *testing.T) {
	packet := types.Packet{
		TunnelID:  1,
		Sequence:  3,
		Prices:    []feedstypes.Price{{SignalID: "signal_01", Price: 2}},
		CreatedAt: 123,
	}

	encodingPacket, err := types.EncodingHyperlaneStride(packet)
	require.NoError(t, err)

	expectedMsg := (
	// Function selector for relayPacket (first 4 bytes of Keccak-256 hash of the function signature)
	"200e1407" +
		// Offset to packet data (32 bytes after function selector and padding)
		"0000000000000000000000000000000000000000000000000000000000000020" +
		// TunnelID: 1 (uint64)
		"0000000000000000000000000000000000000000000000000000000000000001" +
		// Nonce: 3 (uint64)
		"0000000000000000000000000000000000000000000000000000000000000003" +
		// Offset to SignalPrices data (128 bytes from start of packet data)
		"0000000000000000000000000000000000000000000000000000000000000080" +
		// CreatedAt: 123 (int64)
		"000000000000000000000000000000000000000000000000000000000000007b" +
		// Length of SignalPrices array: 1
		"0000000000000000000000000000000000000000000000000000000000000001" +
		// SignalPrices[0].SignalID: "signal_01" (padded to 32 bytes)
		"00000000000000000000000000000000000000000000007369676e616c5f3031" +
		// SignalPrices[0].Price: 2 (uint64)
		"0000000000000000000000000000000000000000000000000000000000000002")

	require.Equal(t, expectedMsg, hex.EncodeToString(encodingPacket))
}
