package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetSetLatestSignalPrices(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
	tunnelID := uint64(1)
	latestSignalPrices := types.LatestSignalPrices{
		TunnelID: tunnelID,
		SignalPrices: []types.SignalPrice{
			{SignalID: "BTC", Price: 50000},
		},
	}

	// Set the latest signal prices
	k.SetLatestSignalPrices(ctx, latestSignalPrices)

	// Get the latest signal prices
	retrievedSignalPrices, err := k.GetLatestSignalPrices(ctx, tunnelID)
	require.NoError(t, err)
	require.Equal(t, latestSignalPrices, retrievedSignalPrices)
}

func TestGetAllLatestSignalPrices(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
	latestSignalPrices1 := types.LatestSignalPrices{
		TunnelID: 1,
		SignalPrices: []types.SignalPrice{
			{SignalID: "BTC", Price: 50000},
		},
	}
	latestSignalPrices2 := types.LatestSignalPrices{
		TunnelID: 2,
		SignalPrices: []types.SignalPrice{
			{SignalID: "ETH", Price: 3000},
		},
	}

	// Set the latest signal prices
	k.SetLatestSignalPrices(ctx, latestSignalPrices1)
	k.SetLatestSignalPrices(ctx, latestSignalPrices2)

	// Get all latest signal prices
	allLatestSignalPrices := k.GetAllLatestSignalPrices(ctx)
	require.Len(t, allLatestSignalPrices, 2)
	require.Contains(t, allLatestSignalPrices, latestSignalPrices1)
	require.Contains(t, allLatestSignalPrices, latestSignalPrices2)
}
