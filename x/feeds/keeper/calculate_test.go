package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestCalculateInterval(t *testing.T) {
	params := types.NewParams("[NOT_SET]", 30, 30, 60, 3600, 1000_000_000, 100, 30, 50, 3000, 28800, 10)

	testCases := []struct {
		name        string
		power       int64
		expInterval int64
	}{
		{
			name:        "power less than threshold",
			power:       10000,
			expInterval: 0,
		},
		{
			name:        "power at the threshold",
			power:       1000000000,
			expInterval: 3600,
		},
		{
			name:        "power at minimum interval",
			power:       60000000000,
			expInterval: 60,
		},
		{
			name:        "power exceed the minimum interval",
			power:       600000000000,
			expInterval: 60,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			interval := CalculateInterval(tc.power, params)
			require.Equal(tt, tc.expInterval, interval)
		})
	}
}

func TestCalculateDeviation(t *testing.T) {
	params := types.NewParams("[NOT_SET]", 30, 30, 60, 3600, 1000_000_000, 100, 30, 50, 3000, 28800, 10)

	testCases := []struct {
		name         string
		power        int64
		expDeviation int64
	}{
		{
			name:         "power less than threshold",
			power:        10000,
			expDeviation: 0,
		},
		{
			name:         "power at the threshold",
			power:        1000000000,
			expDeviation: 3000,
		},
		{
			name:         "power at minimum deviation",
			power:        60000000000,
			expDeviation: 50,
		},
		{
			name:         "power exceed the minimum deviation",
			power:        600000000000,
			expDeviation: 50,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			deviation := CalculateDeviation(tc.power, params)
			require.Equal(tt, tc.expDeviation, deviation)
		})
	}
}

func TestSumPower(t *testing.T) {
	require.Equal(t, int64(1250009), sumPower([]types.Signal{
		{
			ID:    "CS:BAND-USD",
			Power: 100000,
		},
		{
			ID:    "CS:ATOM-USD",
			Power: 150000,
		},
		{
			ID:    "CS:OSMO-USD",
			Power: 1000009,
		},
	}))
}
