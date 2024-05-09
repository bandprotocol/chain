package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestCalculateMedianPriceFeedInfo(t *testing.T) {
	testCases := []struct {
		name           string
		priceFeedInfos []types.PriceFeedInfo
		expRes         uint64
	}{
		{
			name: "case 1",
			priceFeedInfos: []types.PriceFeedInfo{
				{Price: 100, Deviation: 10, Power: 100, Timestamp: 100, Index: 0},
				{Price: 103, Deviation: 10, Power: 100, Timestamp: 101, Index: 1},
				{Price: 105, Deviation: 10, Power: 100, Timestamp: 102, Index: 2},
				{Price: 107, Deviation: 10, Power: 100, Timestamp: 103, Index: 3},
				{Price: 109, Deviation: 10, Power: 100, Timestamp: 104, Index: 4},
			},
			expRes: 107,
		},
		{
			name: "case 2",
			priceFeedInfos: []types.PriceFeedInfo{
				{Price: 100, Deviation: 10, Power: 100, Timestamp: 100, Index: 0},
				{Price: 103, Deviation: 10, Power: 200, Timestamp: 101, Index: 1},
				{Price: 105, Deviation: 10, Power: 300, Timestamp: 102, Index: 2},
				{Price: 107, Deviation: 10, Power: 400, Timestamp: 103, Index: 3},
				{Price: 109, Deviation: 10, Power: 500, Timestamp: 104, Index: 4},
			},
			expRes: 109,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			price, err := types.CalculateMedianPriceFeedInfo(tc.priceFeedInfos)
			require.NoError(tt, err)
			require.Equal(tt, tc.expRes, price)
		})
	}
}

func TestGetDeviationWeightedPrices(t *testing.T) {
	dwp := types.GetDeviationWeightedPrices(100, 5, 1000)
	require.Equal(t, []types.WeightedPrice{
		{
			Power: 1000,
			Price: 100,
		},
		{
			Power: 1000,
			Price: 95,
		},
		{
			Power: 1000,
			Price: 105,
		},
	}, dwp)
}

func TestCalculateMedianWeightedPrice(t *testing.T) {
	testCases := []struct {
		name           string
		weightedPrices []types.WeightedPrice
		expRes         uint64
	}{
		{
			name: "case 1",
			weightedPrices: []types.WeightedPrice{
				{Price: 100, Power: 100},
				{Price: 103, Power: 100},
				{Price: 105, Power: 100},
				{Price: 107, Power: 100},
				{Price: 109, Power: 100},
			},
			expRes: 105,
		},
		{
			name: "case 2",
			weightedPrices: []types.WeightedPrice{
				{Price: 100, Power: 100},
				{Price: 103, Power: 200},
				{Price: 105, Power: 300},
				{Price: 107, Power: 400},
				{Price: 109, Power: 500},
			},
			expRes: 107,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			price, err := types.CalculateMedianWeightedPrice(tc.weightedPrices)
			require.NoError(tt, err)
			require.Equal(tt, tc.expRes, price)
		})
	}
}
