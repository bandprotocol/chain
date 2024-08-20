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
		[]types.SignalPriceInfo{{SignalID: "BTC", DeviationBPS: 1000, Interval: 10, Price: 1000, LastTimestamp: 0}},
	)

	require.Equal(
		t,
		[]byte(
			`{"nonce":"1","signal_price_infos":[{"deviation_bps":"1000","interval":"10","last_timestamp":"0","price":"1000","signal_id":"BTC"}],"tunnel_id":"1"}`,
		),
		packet.GetBytes(),
	)
}
