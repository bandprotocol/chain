package keeper_test

import (
	"fmt"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestAddTunnel() {
	ctx, k := s.ctx, s.keeper

	route := &types.TSSRoute{}
	any, _ := codectypes.NewAnyWithValue(route)
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address"))

	expectedTunnel := types.Tunnel{
		ID:               1,
		Route:            any,
		Encoder:          feedstypes.ENCODER_FIXED_POINT_ABI,
		FeePayer:         "band1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62q2yggu0",
		Creator:          creator.String(),
		Interval:         interval,
		SignalDeviations: signalDeviations,
		IsActive:         false,
		CreatedAt:        ctx.BlockTime().Unix(),
		TotalDeposit:     sdk.NewCoins(),
	}

	expectedPrices := types.LatestPrices{
		TunnelID:     1,
		Prices:       []feedstypes.Price(nil),
		LastInterval: 0,
	}

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	tunnel, err := k.AddTunnel(ctx, route, feedstypes.ENCODER_FIXED_POINT_ABI, signalDeviations, interval, creator)
	s.Require().NoError(err)
	s.Require().Equal(expectedTunnel, *tunnel)

	// check the tunnel count
	tunnelCount := k.GetTunnelCount(ctx)
	s.Require().Equal(uint64(1), tunnelCount)

	// check the latest prices
	latestPrices, err := k.GetLatestPrices(ctx, tunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(expectedPrices, latestPrices)
}

func (s *KeeperTestSuite) TestUpdateAndResetTunnel() {
	ctx, k := s.ctx, s.keeper

	initialRoute := &types.TSSRoute{}
	initialEncoder := feedstypes.ENCODER_FIXED_POINT_ABI
	initialSignalDeviations := []types.SignalDeviation{
		{SignalID: "BTC", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		{SignalID: "ETH", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
	}
	initialInterval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address"))

	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	initialTunnel, err := k.AddTunnel(
		ctx,
		initialRoute,
		initialEncoder,
		initialSignalDeviations,
		initialInterval,
		creator,
	)
	s.Require().NoError(err)

	// define new test data for editing the tunnel
	newSignalDeviations := []types.SignalDeviation{
		{SignalID: "BTC", SoftDeviationBPS: 1100, HardDeviationBPS: 1100},
		{SignalID: "ETH", SoftDeviationBPS: 1100, HardDeviationBPS: 1100},
	}
	newInterval := uint64(20)

	// call the UpdateAndResetTunnel function
	err = k.UpdateAndResetTunnel(ctx, initialTunnel.ID, newSignalDeviations, newInterval)
	s.Require().NoError(err)

	// validate the edited tunnel
	editedTunnel, err := k.GetTunnel(ctx, initialTunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(newSignalDeviations, editedTunnel.SignalDeviations)
	s.Require().Equal(newInterval, editedTunnel.Interval)

	// check the latest prices
	latestPrices, err := k.GetLatestPrices(ctx, editedTunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(editedTunnel.ID, latestPrices.TunnelID)
	s.Require().Len(latestPrices.Prices, 0)
	for i, sp := range latestPrices.Prices {
		s.Require().Equal(newSignalDeviations[i].SignalID, sp.SignalID)
		s.Require().Equal(uint64(0), sp.Price)
	}
}

func (s *KeeperTestSuite) TestGetSetTunnel() {
	ctx, k := s.ctx, s.keeper

	tunnel := types.Tunnel{ID: 1}

	k.SetTunnel(ctx, tunnel)

	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(tunnel, retrievedTunnel)
}

func (s *KeeperTestSuite) TestGetTunnels() {
	ctx, k := s.ctx, s.keeper

	tunnel := types.Tunnel{ID: 1}

	k.SetTunnel(ctx, tunnel)

	tunnels := k.GetTunnels(ctx)
	s.Require().Len(tunnels, 1)
}

func (s *KeeperTestSuite) TestGetSetTunnelCount() {
	ctx, k := s.ctx, s.keeper

	newCount := uint64(5)
	k.SetTunnelCount(ctx, newCount)

	retrievedCount := k.GetTunnelCount(ctx)
	s.Require().Equal(newCount, retrievedCount)
}

func (s *KeeperTestSuite) TestGetActiveTunnelIDs() {
	ctx, k := s.ctx, s.keeper

	activeTunnelIDs := []uint64{1, 2, 3}
	for _, id := range activeTunnelIDs {
		k.ActiveTunnelID(ctx, id)
	}

	retrievedIDs := k.GetActiveTunnelIDs(ctx)
	s.Require().Equal(activeTunnelIDs, retrievedIDs)
}

func (s *KeeperTestSuite) TestActivateTunnel() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	route := &codectypes.Any{}
	encoder := feedstypes.ENCODER_FIXED_POINT_ABI
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	tunnel := types.Tunnel{
		ID:               tunnelID,
		Route:            route,
		Encoder:          encoder,
		SignalDeviations: signalDeviations,
		Interval:         interval,
		TotalDeposit:     k.GetParams(ctx).MinDeposit,
		Creator:          creator,
		IsActive:         false,
		CreatedAt:        ctx.BlockTime().Unix(),
	}

	k.SetTunnel(ctx, tunnel)

	err := k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	// validate the tunnel is activated
	activatedTunnel, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().True(activatedTunnel.IsActive)

	// validate the active tunnel ID is stored
	activeTunnelIDs := k.GetActiveTunnelIDs(ctx)
	s.Require().Contains(activeTunnelIDs, tunnelID)
}

func (s *KeeperTestSuite) TestDeactivateTunnel() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	route := &codectypes.Any{}
	encoder := feedstypes.ENCODER_FIXED_POINT_ABI
	signalDeviations := []types.SignalDeviation{
		{SignalID: "BTC"},
		{SignalID: "ETH"},
	}
	interval := uint64(10)
	creator := sdk.AccAddress([]byte("creator_address")).String()

	// add a tunnel to the store
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

	// call the DeactivateTunnel function
	err := k.DeactivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	// validate the tunnel is deactivated
	deactivatedTunnel, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().False(deactivatedTunnel.IsActive)

	// validate the active tunnel ID is removed
	activeTunnelIDs := k.GetActiveTunnelIDs(ctx)
	s.Require().NotContains(activeTunnelIDs, tunnelID)
}

func (s *KeeperTestSuite) TestGetSetTotalFees() {
	ctx, k := s.ctx, s.keeper

	totalFees := types.TotalFees{TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))}
	k.SetTotalFees(ctx, totalFees)

	retrievedFees := k.GetTotalFees(ctx)
	s.Require().Equal(totalFees, retrievedFees)
}

func (s *KeeperTestSuite) TestGenerateTunnelAccount() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	addr, err := k.GenerateTunnelAccount(ctx, fmt.Sprintf("%d", tunnelID))
	s.Require().NoError(err, "expected no error generating account")
	s.Require().NotNil(addr, "expected generated address to be non-nil")
	s.Require().Equal(
		"band1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62q2yggu0",
		addr.String(),
		"expected generated address to match",
	)
}
