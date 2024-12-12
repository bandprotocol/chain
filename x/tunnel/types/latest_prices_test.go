package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestLatestPrices_UpdatePrices(t *testing.T) {
	initialPrices := []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal1", Price: 100, Timestamp: 1732000000},
	}
	latestPrices := types.NewLatestPrices(1, initialPrices, 10)

	newPrices := []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal1", Price: 200, Timestamp: 1733000000},
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "signal2", Price: 300, Timestamp: 1733000000},
	}
	latestPrices.UpdatePrices(newPrices)

	require.Len(t, latestPrices.Prices, 2)
	require.Equal(t, uint64(200), latestPrices.Prices[0].Price)
	require.Equal(t, uint64(300), latestPrices.Prices[1].Price)
}
