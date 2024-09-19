package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestAddTunnel() {
	ctx, k := s.ctx, s.keeper

	// Define test data
	route := &codectypes.Any{}
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	// Call the AddTunnel function
	tunnel, err := k.AddTunnel(ctx, route, types.ENCODER_FIXED_POINT_ABI, signalDeviations, interval, creator)
	s.Require().NoError(err)

	// Define the expected tunnel
	expectedTunnel := types.Tunnel{
		ID:               1,
		Route:            route,
		Encoder:          types.ENCODER_FIXED_POINT_ABI,
		FeePayer:         "cosmos1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62qh49enj",
		Creator:          creator,
		Interval:         interval,
		SignalDeviations: signalDeviations,
		IsActive:         false,
		CreatedAt:        ctx.BlockTime().Unix(),
	}

	// Define the expected latest signal prices
	expectedSignalPrices := types.LatestSignalPrices{
		TunnelID: 1,
		SignalPrices: []types.SignalPrice{
			{SignalID: "BTC", Price: 0},
			{SignalID: "ETH", Price: 0},
		},
		Timestamp: 0,
	}

	// Validate the results
	s.Require().Equal(expectedTunnel, *tunnel)

	// Check the tunnel count
	tunnelCount := k.GetTunnelCount(ctx)
	s.Require().Equal(uint64(1), tunnelCount)

	// Check the latest signal prices
	latestSignalPrices, err := k.GetLatestSignalPrices(ctx, tunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(expectedSignalPrices, latestSignalPrices)
}

func (s *KeeperTestSuite) TestEditTunnel() {
	ctx, k := s.ctx, s.keeper

	// Define initial test data
	initialRoute := &codectypes.Any{}
	initialEncoder := types.ENCODER_FIXED_POINT_ABI
	initialSignalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	initialInterval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	// Add an initial tunnel
	initialTunnel, err := k.AddTunnel(
		ctx,
		initialRoute,
		initialEncoder,
		initialSignalDeviations,
		initialInterval,
		creator,
	)
	s.Require().NoError(err)

	// Define new test data for editing the tunnel
	newSignalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	newInterval := uint64(20)

	// Call the EditTunnel function
	err = k.EditTunnel(ctx, initialTunnel.ID, newSignalDeviations, newInterval)
	s.Require().NoError(err)

	// Validate the edited tunnel
	editedTunnel, err := k.GetTunnel(ctx, initialTunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(newSignalDeviations, editedTunnel.SignalDeviations)
	s.Require().Equal(newInterval, editedTunnel.Interval)

	// Check the latest signal prices
	latestSignalPrices, err := k.GetLatestSignalPrices(ctx, editedTunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(editedTunnel.ID, latestSignalPrices.TunnelID)
	s.Require().Len(latestSignalPrices.SignalPrices, len(newSignalDeviations))
	for i, sp := range latestSignalPrices.SignalPrices {
		s.Require().Equal(newSignalDeviations[i].SignalID, sp.SignalID)
		s.Require().Equal(uint64(0), sp.Price)
	}
}

func (s *KeeperTestSuite) TestGetSetTunnel() {
	ctx, k := s.ctx, s.keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Attempt to retrieve the tunnel by its ID
	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)

	// Assert no error occurred during retrieval
	s.Require().NoError(err, "retrieving tunnel should not produce an error")

	// Assert the retrieved tunnel matches the one we set
	s.Require().Equal(tunnel, retrievedTunnel, "the retrieved tunnel should match the original")
}

func (s *KeeperTestSuite) TestGetTunnels() {
	ctx, k := s.ctx, s.keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Set the tunnel in the keeper
	k.SetTunnel(ctx, tunnel)

	// Retrieve all tunnels
	tunnels := k.GetTunnels(ctx)

	// Assert the number of tunnels is 1
	s.Require().Len(tunnels, 1, "expected 1 tunnel to be retrieved")
}

func (s *KeeperTestSuite) TestGetSetTunnelCount() {
	ctx, k := s.ctx, s.keeper

	// Set a new tunnel count
	newCount := uint64(5)
	k.SetTunnelCount(ctx, newCount)

	// Get the tunnel count and verify it
	retrievedCount := k.GetTunnelCount(ctx)
	s.Require().Equal(newCount, retrievedCount, "retrieved tunnel count should match the set value")
}

func (s *KeeperTestSuite) TestGetActiveTunnelIDs() {
	ctx, k := s.ctx, s.keeper

	// Define active tunnel IDs
	activeTunnelIDs := []uint64{1, 2, 3}

	// Add active tunnel IDs to the store
	for _, id := range activeTunnelIDs {
		k.ActiveTunnelID(ctx, id)
	}

	// Call the GetActiveTunnelIDs function
	retrievedIDs := k.GetActiveTunnelIDs(ctx)

	// Validate the results
	s.Require().Equal(activeTunnelIDs, retrievedIDs, "retrieved active tunnel IDs should match the expected values")
}

func (s *KeeperTestSuite) TestActivateTunnel() {
	ctx, k := s.ctx, s.keeper

	// Define test data
	tunnelID := uint64(1)
	route := &codectypes.Any{}               // Replace with actual route data
	encoder := types.ENCODER_FIXED_POINT_ABI // Replace with actual encoder data
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	// Add a tunnel to the store
	tunnel := types.Tunnel{
		ID:               tunnelID,
		Route:            route,
		Encoder:          encoder,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		Creator:          creator,
		IsActive:         false,
		CreatedAt:        ctx.BlockTime().Unix(),
	}
	k.SetTunnel(ctx, tunnel)

	// Call the ActivateTunnel function
	err := k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	// Validate the tunnel is activated
	activatedTunnel, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().True(activatedTunnel.IsActive, "tunnel should be active")

	// Validate the active tunnel ID is stored
	activeTunnelIDs := k.GetActiveTunnelIDs(ctx)
	s.Require().Contains(activeTunnelIDs, tunnelID, "active tunnel IDs should contain the activated tunnel ID")
}

func (s *KeeperTestSuite) TestDeactivateTunnel() {
	ctx, k := s.ctx, s.keeper

	// Define test data
	tunnelID := uint64(1)
	route := &codectypes.Any{}               // Replace with actual route data
	encoder := types.ENCODER_FIXED_POINT_ABI // Replace with actual encoder data
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	// Add a tunnel to the store
	tunnel := types.Tunnel{
		ID:               tunnelID,
		Route:            route,
		Encoder:          encoder,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		Creator:          creator,
		IsActive:         true,
		CreatedAt:        ctx.BlockTime().Unix(),
	}
	k.SetTunnel(ctx, tunnel)

	// Call the DeactivateTunnel function
	err := k.DeactivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	// Validate the tunnel is deactivated
	deactivatedTunnel, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().False(deactivatedTunnel.IsActive, "tunnel should be inactive")

	// Validate the active tunnel ID is removed
	activeTunnelIDs := k.GetActiveTunnelIDs(ctx)
	s.Require().
		NotContains(activeTunnelIDs, tunnelID, "active tunnel IDs should not contain the deactivated tunnel ID")
}

func (s *KeeperTestSuite) TestGetSetTotalFees() {
	ctx, k := s.ctx, s.keeper

	totalFees := types.TotalFees{TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100)))}

	// Set the total fees in the keeper
	k.SetTotalFees(ctx, totalFees)

	// Get the total fees and verify it
	retrievedFees := k.GetTotalFees(ctx)
	s.Require().Equal(totalFees, retrievedFees, "retrieved total fees should match the set value")
}
