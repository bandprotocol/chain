package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetSetTunnel(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Attempt to retrieve the tunnel by its ID
	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)

	// Assert no error occurred during retrieval
	require.NoError(s.T(), err, "retrieving tunnel should not produce an error")

	// Assert the retrieved tunnel matches the one we set
	require.Equal(s.T(), tunnel, retrievedTunnel, "the retrieved tunnel should match the original")
}

func TestGetTunnels(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Retrieve all tunnels
	tunnels := k.GetTunnels(ctx)

	// Assert the number of tunnels is 1
	require.Len(s.T(), tunnels, 1, "expected 1 tunnel to be retrieved")
}

func TestGetTunnelsByActiveStatus(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1, IsActive: false}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Retrieve all tunnels by active status
	tunnels := k.GetTunnelsByActiveStatus(ctx, false)

	// Assert the number of active tunnels is 1
	require.Len(s.T(), tunnels, 1, "expected 1 active tunnel to be retrieved")
}

// func TestGenerateSignalPriceInfos(t *testing.T) {
// 	s := testutil.NewTestSuite(t)
// 	ctx := s.Ctx

// 	signalPriceInfos := []types.SignalPriceInfo{
// 		{SignalID: "signal1", SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 		{SignalID: "signal2", SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 	}

// 	latestPricesMap := map[string]feedstypes.Price{
// 		"signal1": {Price: 1000, Timestamp: time.Now().Unix(), PriceStatus: feedstypes.PriceStatusAvailable},
// 		"signal2": {Price: 2000, Timestamp: time.Now().Unix(), PriceStatus: feedstypes.PriceStatusUnavailable},
// 	}

// 	tunnelID := uint64(1)
// 	expected := []types.SignalPriceInfo{
// 		{
// 			SignalID:         "signal1",
// 			SoftDeviationBPS: 0,
// 			HardDeviationBPS: 1000,
// 			Price:            1000,
// 			Timestamp:        latestPricesMap["signal1"].Timestamp,
// 		},
// 		{SignalID: "signal2", SoftDeviationBPS: 0, HardDeviationBPS: 1000, Price: 0, Timestamp: 0},
// 	}

// 	result := keeper.GenerateSignalPriceInfos(ctx, signalPriceInfos, latestPricesMap, tunnelID)

// 	require.Equal(t, expected, result)
// }

// func TestGenerateSignalPriceInfosBasedOnDeviation(t *testing.T) {
// 	s := testutil.NewTestSuite(t)
// 	ctx := s.Ctx

// 	// Define test cases
// 	testCases := []struct {
// 		name             string
// 		signalPriceInfos []types.SignalPriceInfo
// 		latestPricesMap  map[string]feedstypes.Price
// 		tunnelID         uint64
// 		expectedResults  []types.SignalPriceInfo
// 	}{
// 		{
// 			name: "All prices available and within deviation",
// 			signalPriceInfos: []types.SignalPriceInfo{
// 				{SignalID: "signal1", Price: 100, SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 				{SignalID: "signal2", Price: 200, SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 			},
// 			latestPricesMap: map[string]feedstypes.Price{
// 				"signal1": {
// 					SignalID:    "signal1",
// 					Price:       109,
// 					PriceStatus: feedstypes.PriceStatusAvailable,
// 					Timestamp:   1234567890,
// 				},
// 				"signal2": {
// 					SignalID:    "signal2",
// 					Price:       205,
// 					PriceStatus: feedstypes.PriceStatusAvailable,
// 					Timestamp:   1234567891,
// 				},
// 			},
// 			tunnelID:        1,
// 			expectedResults: []types.SignalPriceInfo{},
// 		},
// 		{
// 			name: "Price not available",
// 			signalPriceInfos: []types.SignalPriceInfo{
// 				{SignalID: "signal1", Price: 100, SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 			},
// 			latestPricesMap: map[string]feedstypes.Price{
// 				"signal1": {
// 					SignalID:    "signal1",
// 					Price:       0,
// 					PriceStatus: feedstypes.PriceStatusUnavailable,
// 					Timestamp:   1234567890,
// 				},
// 			},
// 			tunnelID: 1,
// 			expectedResults: []types.SignalPriceInfo{
// 				{SignalID: "signal1", SoftDeviationBPS: 0, HardDeviationBPS: 1000, Price: 0, Timestamp: 0},
// 			},
// 		},
// 		{
// 			name: "Price exceeds hard deviation",
// 			signalPriceInfos: []types.SignalPriceInfo{
// 				{SignalID: "signal1", Price: 100, SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 			},
// 			latestPricesMap: map[string]feedstypes.Price{
// 				"signal1": {
// 					SignalID:    "signal1",
// 					Price:       150,
// 					PriceStatus: feedstypes.PriceStatusAvailable,
// 					Timestamp:   1234567890,
// 				},
// 			},
// 			tunnelID: 1,
// 			expectedResults: []types.SignalPriceInfo{
// 				{
// 					SignalID:         "signal1",
// 					SoftDeviationBPS: 0,
// 					HardDeviationBPS: 1000,
// 					Price:            150,
// 					Timestamp:        1234567890,
// 				},
// 			},
// 		},
// 		{
// 			name: "Signal ID not found",
// 			signalPriceInfos: []types.SignalPriceInfo{
// 				{SignalID: "signal1", Price: 100, SoftDeviationBPS: 0, HardDeviationBPS: 1000},
// 			},
// 			latestPricesMap: map[string]feedstypes.Price{},
// 			tunnelID:        1,
// 			expectedResults: []types.SignalPriceInfo{
// 				{SignalID: "signal1", SoftDeviationBPS: 0, HardDeviationBPS: 1000, Price: 0, Timestamp: 0},
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Call the GenerateSignalPriceInfosBasedOnDeviation method
// 			nsps := keeper.GenerateSignalPriceInfosBasedOnDeviation(
// 				ctx,
// 				tc.signalPriceInfos,
// 				tc.latestPricesMap,
// 				tc.tunnelID,
// 			)

// 			// Verify the results
// 			require.Equal(t, len(tc.expectedResults), len(nsps))
// 			for i, expected := range tc.expectedResults {
// 				require.Equal(t, expected.SignalID, nsps[i].SignalID)
// 				require.Equal(t, expected.Price, nsps[i].Price)
// 				require.Equal(t, expected.Timestamp, nsps[i].Timestamp)
// 			}
// 		})
// 	}
// }

// func TestCreateLatestPricesMap(t *testing.T) {
// 	// Create test data
// 	latestPrices := []feedstypes.Price{
// 		{
// 			SignalID:    "signal1",
// 			Price:       100,
// 			PriceStatus: feedstypes.PriceStatusAvailable,
// 			Timestamp:   1234567890,
// 		},
// 		{
// 			SignalID:    "signal2",
// 			Price:       200,
// 			PriceStatus: feedstypes.PriceStatusAvailable,
// 			Timestamp:   1234567891,
// 		},
// 	}

// 	// Call the createLatestPricesMap method
// 	latestPricesMap := keeper.CreateLatestPricesMap(latestPrices)

// 	// Verify the results
// 	require.Equal(t, 2, len(latestPricesMap))
// 	require.Contains(t, latestPricesMap, "signal1")
// 	require.Contains(t, latestPricesMap, "signal2")
// 	require.Equal(t, latestPrices[0], latestPricesMap["signal1"])
// 	require.Equal(t, latestPrices[1], latestPricesMap["signal2"])
// }
