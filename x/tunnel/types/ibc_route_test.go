package types_test

import (
	b64 "encoding/base64"
	"fmt"
	"testing"

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
		}, "tunnel", "channel-1", 9000000,
	)
	bytes := packet.GetBytes()
	sEnc := b64.StdEncoding.EncodeToString(bytes)
	fmt.Println(sEnc)
	if len(bytes) == 0 {
		t.Errorf("expected non-empty bytes, got empty bytes")
	}
}
