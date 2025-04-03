package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestStringIBCHookMemo(t *testing.T) {
	packet := types.NewTunnelPricesPacketData(
		1,
		2,
		[]feedstypes.Price{
			{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal1", Price: 200, Timestamp: 1740131933},
			{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal2", Price: 300, Timestamp: 1740131933},
		},
		1610000000,
	)

	memo := types.NewIBCHookMemo(
		"wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk",
		packet,
	)
	memoStr := memo.JSONString()
	require.Equal(
		t,
		`{"wasm":{"contract":"wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk","msg":{"receive_packet":{"packet":{"created_at":"1610000000","prices":[{"price":"200","signal_id":"signal1","status":"PRICE_STATUS_AVAILABLE","timestamp":"1740131933"},{"price":"300","signal_id":"signal2","status":"PRICE_STATUS_AVAILABLE","timestamp":"1740131933"}],"sequence":"2","tunnel_id":"1"}}}}}`,
		memoStr,
	)
}
