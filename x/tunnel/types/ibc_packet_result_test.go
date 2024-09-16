package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetByteIBCPacket(t *testing.T) {
	packet := types.NewIBCPacketResult(
		1,
		1,
		[]types.SignalPrice{{SignalID: "BTC", Price: 1000}},
		0,
	)

	require.Equal(
		t,
		[]byte(
			`{"created_at":"0","nonce":"1","signal_prices":[{"price":"1000","signal_id":"BTC"}],"tunnel_id":"1"}`,
		),
		packet.GetBytes(),
	)
}
