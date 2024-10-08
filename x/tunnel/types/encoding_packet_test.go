package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestEncodeABI(t *testing.T) {
	packet := types.Packet{
		TunnelID:     1,
		Sequence:     3,
		SignalPrices: []types.SignalPrice{{SignalID: "signal_01", Price: 2}},
		CreatedAt:    123,
	}

	encodingPacket, err := types.NewEncodingPacket(packet, types.ENCODER_FIXED_POINT_ABI)
	require.NoError(t, err)

	msg, err := encodingPacket.EncodeABI()
	require.NoError(t, err)

	expectedMsg := ("0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000080" +
		"000000000000000000000000000000000000000000000000000000000000007b" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"00000000000000000000000000000000000000000000007369676e616c5f3031" +
		"0000000000000000000000000000000000000000000000000000000000000002")

	require.Equal(t, expectedMsg, hex.EncodeToString(msg))
}
