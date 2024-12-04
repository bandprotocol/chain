package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetByteTunnelPricesPacketData(t *testing.T) {
	packet := types.NewTunnelPricesPacketData(
		1,
		2,
		[]feedstypes.Price{
			{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 50000, Timestamp: 1733000000},
		},
		1633024800,
	)

	require.Equal(
		t,
		[]byte(
			`{"created_at":"1633024800","prices":[{"price":"50000","signal_id":"CS:BAND-USD","status":"PRICE_STATUS_AVAILABLE","timestamp":"1733000000"}],"sequence":"2","tunnel_id":"1"}`,
		),
		packet.GetBytes(),
	)
}
