package tunnel_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestPacketDataUnmarshalerInterface(t *testing.T) {
	var (
		data          []byte
		expPacketData types.TunnelPricesPacketData
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"invalid packet data",
			func() {
				data = []byte("invalid packet data")
			},
			false,
		},
		{
			"all good",
			func() {
				expPacketData = types.TunnelPricesPacketData{
					TunnelID: 1,
					Sequence: 1,
					Prices: []feedstypes.Price{
						{
							Status:    feedstypes.PRICE_STATUS_AVAILABLE,
							SignalID:  "CS:BAND-USD",
							Price:     50000,
							Timestamp: 1733000000,
						},
					},
					CreatedAt: 1633024800,
				}
				data = expPacketData.GetBytes()
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.malleate()

			packetData, err := tunnel.IBCModule{}.UnmarshalPacketData(data)

			if tc.expPass {
				require.NoError(t, err)
				require.Equal(t, expPacketData, packetData)
			} else {
				require.Error(t, err)
				require.Nil(t, packetData)
			}
		})
	}
}
