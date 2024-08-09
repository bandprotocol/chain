package types_test

import (
	b64 "encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetIBCPacketBytes(t *testing.T) {
	packet := types.NewIBCPacket(
		1, 1, 1, []types.SignalPriceInfo{
			{
				SignalID:      "BTC",
				DeviationBPS:  1000,
				Interval:      10,
				Price:         1000,
				LastTimestamp: 0,
			},
		}, "channel-1", 9000000,
	)
	bytes := packet.GetBytes()
	sEnc := b64.StdEncoding.EncodeToString(bytes)
	fmt.Println(sEnc)
	if len(bytes) == 0 {
		t.Errorf("expected non-empty bytes, got empty bytes")
	}
}

func TestGetBytesResponsePacket(t *testing.T) {
	packet := types.NewIBCPacket(
		1, 1, 1, []types.SignalPriceInfo{
			{
				SignalID:      "BTC",
				DeviationBPS:  1000,
				Interval:      10,
				Price:         1000,
				LastTimestamp: 0,
			},
		}, "channel-1", 9000000,
	)

	fmt.Println(string(packet.GetBytes()))

	require.Equal(
		t,
		[]byte(
			`{"channel_id":"channel-1","created_at":"9000000","feed_type":"FEED_TYPE_DEFAULT","nonce":"1","signal_price_infos":[{"deviation_bps":"1000","interval":"10","last_timestamp":"0","price":"1000","signal_id":"BTC"}],"tunnel_id":"1"}`,
		),
		packet.GetBytes(),
	)
}
