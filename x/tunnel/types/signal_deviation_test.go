package types_test

import (
	"testing"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// mockContext returns a sdk.Context with the given chainID and blockHeight
func mockContext(chainID string, blockHeight int64) sdk.Context {
	header := cmtproto.Header{ChainID: chainID, Height: blockHeight}
	return sdk.NewContext(nil, header, false, nil)
}

func TestValidateSignalDeviations(t *testing.T) {
	params := types.Params{
		MaxSignals:      5,
		MinDeviationBPS: 10,
		MaxDeviationBPS: 100,
	}

	testCases := []struct {
		name             string
		ctx              sdk.Context
		signalDeviations []types.SignalDeviation
		expErr           bool
		expErrMsg        string
	}{
		{
			name: "legacy: too many signals",
			ctx:  mockContext("otherchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 20, SoftDeviationBPS: 30},
				{SignalID: "CS:ETH-USD", HardDeviationBPS: 40, SoftDeviationBPS: 50},
				{SignalID: "CS:XRP-USD", HardDeviationBPS: 60, SoftDeviationBPS: 70},
				{SignalID: "CS:LTC-USD", HardDeviationBPS: 80, SoftDeviationBPS: 90},
				{SignalID: "CS:BCH-USD", HardDeviationBPS: 100, SoftDeviationBPS: 110},
				{SignalID: "CS:ADA-USD", HardDeviationBPS: 120, SoftDeviationBPS: 130},
			},
			expErr:    true,
			expErrMsg: "max signals 5, got 6",
		},
		{
			name: "legacy: deviation too low",
			ctx:  mockContext("otherchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 5, SoftDeviationBPS: 5},
			},
			expErr:    true,
			expErrMsg: "min 10, max 100, got 5, 5",
		},
		{
			name: "legacy: deviation too high",
			ctx:  mockContext("otherchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 150, SoftDeviationBPS: 150},
			},
			expErr:    true,
			expErrMsg: "min 10, max 100, got 150, 150",
		},
		{
			name: "legacy: all good",
			ctx:  mockContext("otherchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 30, SoftDeviationBPS: 20},
				{SignalID: "CS:ETH-USD", HardDeviationBPS: 50, SoftDeviationBPS: 40},
			},
			expErr:    false,
			expErrMsg: "",
		},
		{
			name: "new: soft deviation greater than hard deviation",
			ctx:  mockContext("bandchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 30, SoftDeviationBPS: 40},
			},
			expErr:    true,
			expErrMsg: "got 40, 30",
		},
		{
			name: "new: all good",
			ctx:  mockContext("bandchain", 1),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 30, SoftDeviationBPS: 20},
				{SignalID: "CS:ETH-USD", HardDeviationBPS: 50, SoftDeviationBPS: 40},
				{SignalID: "CS:BAND-USD", HardDeviationBPS: 50, SoftDeviationBPS: 0},
			},
			expErr:    false,
			expErrMsg: "",
		},
		{
			name: "band-v3-testnet-1: legacy logic (height below threshold)",
			ctx:  mockContext("band-v3-testnet-1", 39836000),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 15, SoftDeviationBPS: 0},
			},
			expErr:    true,
			expErrMsg: "min 10, max 100, got 0, 15",
		},
		{
			name: "band-v3-testnet-1: new logic (height above threshold)",
			ctx:  mockContext("band-v3-testnet-1", 39836001),
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 15, SoftDeviationBPS: 0},
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateSignalDeviations(
				tc.ctx,
				tc.signalDeviations,
				params.MaxSignals,
				params.MaxDeviationBPS,
				params.MinDeviationBPS,
			)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsDeviationOutOfRange(t *testing.T) {
	cases := []struct {
		name string
		sd   types.SignalDeviation
		min  uint64
		max  uint64
		want bool
	}{
		{"hard below min", types.SignalDeviation{HardDeviationBPS: 5, SoftDeviationBPS: 0}, 10, 100, true},
		{"hard above max", types.SignalDeviation{HardDeviationBPS: 150, SoftDeviationBPS: 0}, 10, 100, true},
		{"soft above hard", types.SignalDeviation{HardDeviationBPS: 20, SoftDeviationBPS: 30}, 10, 100, true},
		{"all good", types.SignalDeviation{HardDeviationBPS: 30, SoftDeviationBPS: 0}, 10, 100, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := types.IsDeviationOutOfRange(tc.sd, tc.max, tc.min)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsDeviationOutOfRangeLegacy(t *testing.T) {
	cases := []struct {
		name string
		sd   types.SignalDeviation
		min  uint64
		max  uint64
		want bool
	}{
		{"hard below min", types.SignalDeviation{HardDeviationBPS: 5, SoftDeviationBPS: 20}, 10, 100, true},
		{"soft below min", types.SignalDeviation{HardDeviationBPS: 20, SoftDeviationBPS: 5}, 10, 100, true},
		{"hard above max", types.SignalDeviation{HardDeviationBPS: 150, SoftDeviationBPS: 20}, 10, 100, true},
		{"soft above max", types.SignalDeviation{HardDeviationBPS: 20, SoftDeviationBPS: 150}, 10, 100, true},
		{"all good", types.SignalDeviation{HardDeviationBPS: 30, SoftDeviationBPS: 20}, 10, 100, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := types.IsDeviationOutOfRangeLegacy(tc.sd, tc.max, tc.min)
			require.Equal(t, tc.want, got)
		})
	}
}
