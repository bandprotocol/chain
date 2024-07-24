package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestAddTunnel(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Mock the account keeper to generate a new account
	s.MockAccountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.MockAccountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.MockAccountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	// Add the tunnel to the keeper
	_, err := k.AddTunnel(ctx, tunnel)
	require.NoError(s.T(), err, "adding tunnel should not produce an error")

	// Attempt to retrieve the tunnel by its ID
	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)
	require.NoError(s.T(), err, "retrieving tunnel should not produce an error")

	expected := types.Tunnel{
		ID:                       1,
		Route:                    nil,
		FeedType:                 feedstypes.FEED_TYPE_UNSPECIFIED,
		FeePayer:                 "cosmos1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62qh49enj",
		SignalPriceInfos:         nil,
		LastTriggeredBlockHeight: 0,
		IsActive:                 false,
		CreatedAt:                s.Ctx.BlockTime(),
		Creator:                  "",
	}

	// Assert the retrieved tunnel matches the one we set
	require.Equal(s.T(), expected, retrievedTunnel, "the retrieved tunnel should match the original")
}

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

func TestGetActiveTunnels(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1, IsActive: true}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Retrieve all active tunnels
	tunnels := k.GetActiveTunnels(ctx)

	// Assert the number of active tunnels is 1
	require.Len(s.T(), tunnels, 1, "expected 1 active tunnel to be retrieved")
}

func TestGetRequiredProcessTunnels(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	now := ctx.BlockTime()
	before := now.Add(-30 * time.Second)

	// Mock data for the test cases
	tunnels := []types.Tunnel{
		{
			ID: 1,
			SignalPriceInfos: []types.SignalPriceInfo{
				{
					SignalID:      "signal1",
					Price:         100,
					DeviationBPS:  1000,
					Interval:      30,
					LastTimestamp: &now,
				},
			},
			IsActive: true,
		},
		{
			ID: 2,
			SignalPriceInfos: []types.SignalPriceInfo{
				{SignalID: "signal2", Price: 100, DeviationBPS: 1000, Interval: 30, LastTimestamp: &now},
			},
			IsActive: true,
		},
		{
			ID: 3,
			SignalPriceInfos: []types.SignalPriceInfo{
				{SignalID: "signal3", Price: 100, DeviationBPS: 1000, Interval: 30, LastTimestamp: &before},
			},
			IsActive: true,
		},
	}
	prices := []feedstypes.Price{
		{SignalID: "signal1", Price: 110},
		{SignalID: "signal2", Price: 111},
		{SignalID: "signal3", Price: 101},
	}

	for _, tunnel := range tunnels {
		k.SetTunnel(ctx, tunnel)
	}
	s.MockFeedsKeeper.EXPECT().GetPrices(ctx).Return(prices).Times(1)

	// Execute the function to test
	resultTunnels := k.GetRequiredProcessTunnels(ctx)

	// Assert conditions
	require.Len(t, resultTunnels, 2, "There should be 2 tunnels requiring processing")
	require.Equal(t, uint64(2), resultTunnels[0].ID, "The tunnel requiring processing should be tunnel1")

	// check for correct updates in the SignalPriceInfos
	require.Equal(
		t,
		uint64(111),
		resultTunnels[0].SignalPriceInfos[0].Price,
		"The price should be updated to the latest price",
	)
	require.Equal(
		t,
		uint64(101),
		resultTunnels[1].SignalPriceInfos[0].Price,
		"The price should be updated to the latest price",
	)
}

func TestGetNextTunnelID(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	firstID := k.GetNextTunnelID(ctx)
	require.Equal(s.T(), uint64(1), firstID, "expected first tunnel ID to be 1")

	secondID := k.GetNextTunnelID(ctx)
	require.Equal(s.T(), uint64(2), secondID, "expected next tunnel ID to be 2")
}