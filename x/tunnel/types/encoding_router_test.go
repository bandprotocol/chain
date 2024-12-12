package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestEncodingRouter(t *testing.T) {
	packet := types.Packet{
		TunnelID:  1,
		Sequence:  1,
		Prices:    []feedstypes.Price{{SignalID: "BTC", Price: 72163}},
		CreatedAt: 1730358471,
	}

	encodingPacket, err := types.EncodingRouter(packet)
	require.NoError(t, err)

	expectedMsg := (
	// Offset to packet data (32 bytes after function selector and padding)
	"0000000000000000000000000000000000000000000000000000000000000020" +
		// TunnelID: 1 (uint64)
		"0000000000000000000000000000000000000000000000000000000000000001" +
		// Nonce: 1 (uint64)
		"0000000000000000000000000000000000000000000000000000000000000001" +
		// Offset to SignalPrices data (128 bytes from start of packet data)
		"0000000000000000000000000000000000000000000000000000000000000080" +
		// CreatedAt: 1730358471 (int64)
		"0000000000000000000000000000000000000000000000000000000067232cc7" +
		// Length of SignalPrices array: 1
		"0000000000000000000000000000000000000000000000000000000000000001" +
		// SignalPrices[0].SignalID: "BTC" (padded to 32 bytes)
		"0000000000000000000000000000000000000000000000000000000000425443" +
		// SignalPrices[0].Price: 72163 (uint64)
		"00000000000000000000000000000000000000000000000000000000000119e3")

	require.Equal(t, expectedMsg, hex.EncodeToString(encodingPacket))
}
