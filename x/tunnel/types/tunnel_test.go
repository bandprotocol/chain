package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestGetSetRoute(t *testing.T) {
	tunnel := types.Tunnel{}
	route := &types.TSSRoute{DestinationChainID: "chain-1", DestinationContractAddress: "contract-1"}

	err := tunnel.SetRoute(route)
	require.NoError(t, err)

	routeValue, err := tunnel.GetRouteValue()
	require.NoError(t, err)
	require.Equal(t, route, routeValue)
}

func TestGetSignalDeviationMap(t *testing.T) {
	signalDeviations := []types.SignalDeviation{{SignalID: "signal1", SoftDeviationBPS: 100, HardDeviationBPS: 200}}
	tunnel := types.Tunnel{SignalDeviations: signalDeviations}

	signalDeviationMap := tunnel.GetSignalDeviationMap()
	require.Len(t, signalDeviationMap, 1)
	require.Equal(t, signalDeviations[0], signalDeviationMap["signal1"])
}

func TestGetSignalIDs(t *testing.T) {
	signalDeviations := []types.SignalDeviation{
		{SignalID: "signal1", SoftDeviationBPS: 100, HardDeviationBPS: 200},
		{SignalID: "signal2", SoftDeviationBPS: 100, HardDeviationBPS: 200},
	}
	tunnel := types.Tunnel{SignalDeviations: signalDeviations}

	signalIDs := tunnel.GetSignalIDs()
	require.Len(t, signalIDs, 2)
	require.Contains(t, signalIDs, "signal1")
	require.Contains(t, signalIDs, "signal2")
}

func TestValidateInterval(t *testing.T) {
	tests := []struct {
		name        string
		interval    uint64
		maxInterval uint64
		minInterval uint64
		expErr      bool
	}{
		{
			name:        "interval too low",
			interval:    5,
			maxInterval: 100,
			minInterval: 10,
			expErr:      true,
		},
		{
			name:        "interval too high",
			interval:    150,
			maxInterval: 100,
			minInterval: 10,
			expErr:      true,
		},
		{
			name:        "all good",
			interval:    50,
			maxInterval: 100,
			minInterval: 10,
			expErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := types.ValidateInterval(tt.interval, tt.maxInterval, tt.minInterval)
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
