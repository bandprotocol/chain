package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestLatestSignalPrices_Validate(t *testing.T) {
	cases := map[string]struct {
		latestSignalPrices types.LatestSignalPrices
		expErr             bool
		expErrMsg          string
	}{
		"valid latest signal prices": {
			latestSignalPrices: types.NewLatestSignalPrices(
				1,
				[]types.SignalPrice{{SignalID: "signal1", Price: 100}},
				10,
			),
			expErr: false,
		},
		"invalid tunnel ID": {
			latestSignalPrices: types.NewLatestSignalPrices(
				0,
				[]types.SignalPrice{{SignalID: "signal1", Price: 100}},
				10,
			),
			expErr:    true,
			expErrMsg: "tunnel ID cannot be 0",
		},
		"empty signal prices": {
			latestSignalPrices: types.NewLatestSignalPrices(1, []types.SignalPrice{}, 10),
			expErr:             true,
			expErrMsg:          "signal prices cannot be empty",
		},
		"negative last interval": {
			latestSignalPrices: types.NewLatestSignalPrices(
				1,
				[]types.SignalPrice{{SignalID: "signal1", Price: 100}},
				-1,
			),
			expErr:    true,
			expErrMsg: "last interval cannot be negative",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.latestSignalPrices.Validate()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLatestSignalPrices_UpdateSignalPrices(t *testing.T) {
	initialSignalPrices := []types.SignalPrice{{SignalID: "signal1", Price: 100}}
	latestSignalPrices := types.NewLatestSignalPrices(1, initialSignalPrices, 10)

	newSignalPrices := []types.SignalPrice{{SignalID: "signal1", Price: 200}, {SignalID: "signal2", Price: 300}}
	latestSignalPrices.UpdateSignalPrices(newSignalPrices)

	require.Len(t, latestSignalPrices.SignalPrices, 1)
	require.Equal(t, uint64(200), latestSignalPrices.SignalPrices[0].Price)
}
