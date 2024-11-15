package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestMedianValidatorPriceInfo(t *testing.T) {
	testCases := []struct {
		name                string
		validatorPriceInfos []types.ValidatorPriceInfo
		expRes              uint64
	}{
		{
			name: "case 1",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             100,
					Power:             sdkmath.NewInt(100),
					Timestamp:         100,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             103,
					Power:             sdkmath.NewInt(100),
					Timestamp:         101,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             105,
					Power:             sdkmath.NewInt(100),
					Timestamp:         102,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             107,
					Power:             sdkmath.NewInt(100),
					Timestamp:         103,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             109,
					Power:             sdkmath.NewInt(100),
					Timestamp:         104,
				},
			},
			expRes: 107,
		},
		{
			name: "case 2",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             100,
					Power:             sdkmath.NewInt(100),
					Timestamp:         100,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             103,
					Power:             sdkmath.NewInt(200),
					Timestamp:         101,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             105,
					Power:             sdkmath.NewInt(300),
					Timestamp:         102,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             107,
					Power:             sdkmath.NewInt(400),
					Timestamp:         103,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             109,
					Power:             sdkmath.NewInt(500),
					Timestamp:         104,
				},
			},
			expRes: 109,
		},
		{
			name: "case 3",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             1000,
					Power:             sdkmath.NewInt(5000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             2000,
					Power:             sdkmath.NewInt(4000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             2000,
					Power:             sdkmath.NewInt(4000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             2000,
					Power:             sdkmath.NewInt(4000),
					Timestamp:         1716448424,
				},
			},
			expRes: 1000,
		},
		{
			name: "case 4",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_UNAVAILABLE,
					Price:             0,
					Power:             sdkmath.NewInt(5000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_UNSUPPORTED,
					Price:             0,
					Power:             sdkmath.NewInt(4000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             1000,
					Power:             sdkmath.NewInt(3000),
					Timestamp:         1716448424,
				},
				{
					SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
					Price:             3000,
					Power:             sdkmath.NewInt(4000),
					Timestamp:         1716448424,
				},
			},
			expRes: 3000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			price, err := types.MedianValidatorPriceInfos(tc.validatorPriceInfos)
			require.NoError(tt, err)
			require.Equal(tt, tc.expRes, price)
		})
	}
}

func TestMedianWeightedPrice(t *testing.T) {
	testCases := []struct {
		name           string
		weightedPrices []types.WeightedPrice
		expRes         uint64
	}{
		{
			name: "case 1",
			weightedPrices: []types.WeightedPrice{
				{Price: 100, Weight: sdkmath.NewInt(100)},
				{Price: 103, Weight: sdkmath.NewInt(100)},
				{Price: 105, Weight: sdkmath.NewInt(100)},
				{Price: 107, Weight: sdkmath.NewInt(100)},
				{Price: 109, Weight: sdkmath.NewInt(100)},
			},
			expRes: 105,
		},
		{
			name: "case 2",
			weightedPrices: []types.WeightedPrice{
				{Price: 100, Weight: sdkmath.NewInt(100)},
				{Price: 103, Weight: sdkmath.NewInt(200)},
				{Price: 105, Weight: sdkmath.NewInt(300)},
				{Price: 107, Weight: sdkmath.NewInt(400)},
				{Price: 109, Weight: sdkmath.NewInt(500)},
			},
			expRes: 107,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			price, err := types.MedianWeightedPrice(tc.weightedPrices)
			require.NoError(tt, err)
			require.Equal(tt, tc.expRes, price)
		})
	}
}
