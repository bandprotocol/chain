package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestStringIBCHookMemo(t *testing.T) {
	memo := types.NewIBCHookMemo(
		"wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk",
		1,
		2,
		[]feedstypes.Price{
			{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal1", Price: 200},
			{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal2", Price: 300},
		},
		1610000000,
	)
	memoStr, err := memo.String()
	require.NoError(t, err)
	require.Equal(
		t,
		`{"wasm":{"contract":"wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk","msg":{"receive_band_data":{"tunnel_id":1,"sequence":2,"prices":[{"status":3,"signal_id":"signal1","price":200},{"status":3,"signal_id":"signal2","price":300}],"created_at":1610000000}}}}`,
		memoStr,
	)
}
