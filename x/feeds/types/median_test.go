package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestCalculateMedianPriceFeedInfo(t *testing.T) {
	pfInfos := []types.PriceFeedInfo{
		{
			Power:     1,
			Price:     1,
			Deviation: 0,
			Timestamp: 1,
		},
		{
			Power:     2,
			Price:     2,
			Deviation: 0,
			Timestamp: 2,
		},
	}

	price, err := types.CalculateMedianPriceFeedInfo(pfInfos)
	require.NoError(t, err)
	require.Equal(t, uint64(2), price)
}
