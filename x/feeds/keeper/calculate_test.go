package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestCalculateIntervalAndDeviation(t *testing.T) {
	params := types.NewParams("[NOT_SET]", 30, 30, 60, 3600, 1000_000_000, 100, 30, 5, 300, 256, 28800)

	testCases := []struct {
		name         string
		power        int64
		expInterval  int64
		expDeviation int64
	}{
		{
			name:         "power less than threshold",
			power:        10000,
			expInterval:  0,
			expDeviation: 0,
		},
		{
			name:         "power at the threshold",
			power:        1000000000,
			expInterval:  3600,
			expDeviation: 300,
		},
		{
			name:         "power at minimum interval",
			power:        60000000000,
			expInterval:  60,
			expDeviation: 5,
		},
		{
			name:         "power exceed the minimum interval",
			power:        600000000000,
			expInterval:  60,
			expDeviation: 5,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			interval, deviation := CalculateIntervalAndDeviation(tc.power, params)
			require.Equal(tt, tc.expInterval, interval)
			require.Equal(tt, tc.expDeviation, deviation)
		})
	}
}

func TestSumPower(t *testing.T) {
	require.Equal(t, int64(300000), sumPower([]types.Signal{
		{
			ID:    "crypto_price.bandusd",
			Power: 100000,
		},
		{
			ID:    "crypto_price.atomusd",
			Power: 100000,
		},
		{
			ID:    "crypto_price.osmousd",
			Power: 100000,
		},
	}))
}
