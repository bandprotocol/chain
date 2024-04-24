package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestCalculateMedianPriceFeedInfo(t *testing.T) {
	pfInfos := []types.PriceFeedInfo{
		{Price: 100, Deviation: 10, Power: 100, Timestamp: 100, Index: 0},
		{Price: 103, Deviation: 10, Power: 100, Timestamp: 101, Index: 1},
		{Price: 105, Deviation: 10, Power: 100, Timestamp: 102, Index: 2},
		{Price: 107, Deviation: 10, Power: 100, Timestamp: 103, Index: 3},
		{Price: 109, Deviation: 10, Power: 100, Timestamp: 104, Index: 4},
	}

	price, _ := types.CalculateMedianPriceFeedInfo(pfInfos)
	require.Equal(t, uint64(107), price)
}
