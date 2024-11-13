package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestLatestSignalPrices_Validate(t *testing.T) {
	examplePrices := []feedstypes.Price{
		feedstypes.NewPrice(feedstypes.PriceStatusAvailable, "signal1", 100, 0),
	}

	cases := map[string]struct {
		latestPrices types.LatestPrices
		expErr       bool
		expErrMsg    string
	}{
		"valid latest prices": {
			latestPrices: types.NewLatestPrices(1, examplePrices, 10),
			expErr:       false,
		},
		"invalid tunnel ID": {
			latestPrices: types.NewLatestPrices(0, examplePrices, 10),
			expErr:       true,
			expErrMsg:    "tunnel ID cannot be 0",
		},
		"negative last interval": {
			latestPrices: types.NewLatestPrices(1, examplePrices, -1),
			expErr:       true,
			expErrMsg:    "last interval cannot be negative",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.latestPrices.Validate()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLatestPrices_UpdatePrices(t *testing.T) {
	initialPrices := []feedstypes.Price{
		{Status: feedstypes.PriceStatusAvailable, SignalID: "signal1", Price: 100},
	}
	latestPrices := types.NewLatestPrices(1, initialPrices, 10)

	newPrices := []feedstypes.Price{
		{Status: feedstypes.PriceStatusAvailable, SignalID: "signal1", Price: 200},
		{Status: feedstypes.PriceStatusAvailable, SignalID: "signal2", Price: 300},
	}
	latestPrices.UpdatePrices(newPrices)

	require.Len(t, latestPrices.Prices, 2)
	require.Equal(t, uint64(200), latestPrices.Prices[0].Price)
	require.Equal(t, uint64(300), latestPrices.Prices[1].Price)
}
