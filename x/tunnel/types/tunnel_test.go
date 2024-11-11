package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestTunnel_SetRoute(t *testing.T) {
	tunnel := types.Tunnel{}
	route := &types.TSSRoute{DestinationChainID: "chain-1", DestinationContractAddress: "contract-1"}

	err := tunnel.SetRoute(route)
	require.NoError(t, err)
	require.Equal(t, route, tunnel.Route.GetCachedValue())
}

func TestTunnel_GetSignalDeviationMap(t *testing.T) {
	signalDeviations := []types.SignalDeviation{{SignalID: "signal1", SoftDeviationBPS: 100, HardDeviationBPS: 200}}
	tunnel := types.Tunnel{SignalDeviations: signalDeviations}

	signalDeviationMap := tunnel.GetSignalDeviationMap()
	require.Len(t, signalDeviationMap, 1)
	require.Equal(t, signalDeviations[0], signalDeviationMap["signal1"])
}

func TestTunnel_GetSignalIDs(t *testing.T) {
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

func TestValidateEncoder(t *testing.T) {
	// validate encoder
	err := types.ValidateEncoder(1)
	require.NoError(t, err)

	// invalid encoder
	err = types.ValidateEncoder(999)
	require.Error(t, err)
}
