package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestValidateSignalDeviations(t *testing.T) {
	params := types.Params{
		MaxSignals:      5,
		MinDeviationBPS: 10,
		MaxDeviationBPS: 100,
	}

	tests := []struct {
		name             string
		signalDeviations []types.SignalDeviation
		expErr           bool
		expErrMsg        string
	}{
		{
			name: "too many signals",
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
			name: "deviation too low",
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 5, SoftDeviationBPS: 5},
			},
			expErr:    true,
			expErrMsg: "min 10, max 100, got 5, 5",
		},
		{
			name: "deviation too high",
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 150, SoftDeviationBPS: 150},
			},
			expErr:    true,
			expErrMsg: "min 10, max 100, got 150, 150",
		},
		{
			name: "all good",
			signalDeviations: []types.SignalDeviation{
				{SignalID: "CS:BTC-USD", HardDeviationBPS: 20, SoftDeviationBPS: 30},
				{SignalID: "CS:ETH-USD", HardDeviationBPS: 40, SoftDeviationBPS: 50},
			},
			expErr:    false,
			expErrMsg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateSignalDeviations(
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
