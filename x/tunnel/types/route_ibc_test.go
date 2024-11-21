package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetByteIBCPacket(t *testing.T) {
	packet := types.NewIBCPacketResult(
		1,
		1,
		[]feedstypes.Price{{SignalID: "BTC", Price: 1000}},
		0,
	)

	require.Equal(
		t,
		[]byte(
			`{"created_at":"0","prices":[{"price":"1000","signal_id":"BTC","status":"PRICE_STATUS_UNSPECIFIED","timestamp":"0"}],"sequence":"1","tunnel_id":"1"}`,
		),
		packet.GetBytes(),
	)
}
