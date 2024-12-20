package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestDeductBasePacketFee() {
	ctx, k := s.ctx, s.keeper

	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	basePacketFee := sdk.Coins{sdk.NewInt64Coin("uband", 100)}

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, basePacketFee).
		Return(nil)

	defaultParams := types.DefaultParams()
	defaultParams.BasePacketFee = basePacketFee

	err := k.SetParams(ctx, defaultParams)
	s.Require().NoError(err)

	err = k.DeductBasePacketFee(ctx, feePayer)
	s.Require().NoError(err)

	// validate the total fees are updated
	totalFee := k.GetTotalFees(ctx)
	s.Require().Equal(basePacketFee, totalFee.TotalBasePacketFee)
}

func (s *KeeperTestSuite) TestGetSetPacket() {
	ctx, k := s.ctx, s.keeper

	packet := types.Packet{
		TunnelID: 1,
		Sequence: 1,
	}

	k.SetPacket(ctx, packet)

	storedPacket, err := k.GetPacket(ctx, packet.TunnelID, packet.Sequence)
	s.Require().NoError(err)
	s.Require().Equal(packet, storedPacket)
}

func (s *KeeperTestSuite) TestCreatePacket() {
	ctx, k := s.ctx, s.keeper

	params := k.GetParams(ctx)

	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	tunnel := types.Tunnel{
		ID:       1,
		FeePayer: feePayer.String(),
		IsActive: true,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "0x",
		DestinationContractAddress: "0x",
	}
	prices := []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 5000000, Timestamp: 1733000000},
	}

	s.bandtssKeeper.EXPECT().GetSigningFee(gomock.Any()).Return(
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil,
	)

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, k.GetParams(ctx).BasePacketFee).
		Return(nil)

	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	// set tunnel to the store
	k.SetTunnel(ctx, tunnel)

	expectedPacket := types.Packet{
		TunnelID:  1,
		Sequence:  1,
		Prices:    prices,
		BaseFee:   params.BasePacketFee,
		RouteFee:  sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
		CreatedAt: ctx.BlockTime().Unix(),
	}

	// create a packet
	packet, err := k.CreatePacket(ctx, tunnel.ID, prices)
	s.Require().NoError(err)
	s.Require().Equal(expectedPacket, packet)

	// Verify that the tunnel sequence was incremented
	updatedTunnel, err := k.GetTunnel(ctx, tunnel.ID)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), updatedTunnel.Sequence)
}

func (s *KeeperTestSuite) TestProducePacket() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
		"CS:BAND-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     50000,
			Timestamp: 1733000000,
		},
	}
	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	tunnel := types.Tunnel{
		ID:       1,
		FeePayer: feePayer.String(),
		IsActive: true,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, k.GetParams(ctx).BasePacketFee).
		Return(nil)
	s.bandtssKeeper.EXPECT().GetSigningFee(gomock.Any()).Return(
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil,
	)
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"chain-1",
		"0x1234567890abcdef",
		gomock.Any(),
		feePayer,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).Return(bandtsstypes.SigningID(1), nil)

	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	// set deposit to the tunnel to be able to activate
	tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)

	k.SetTunnel(ctx, tunnel)

	s.bandtssKeeper.EXPECT().IsReady(gomock.Any()).Return(true)
	err = k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	k.SetLatestPrices(ctx, types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 600000, Timestamp: 1732000000},
	}, 0))

	err = k.ProducePacket(ctx, tunnelID, pricesMap)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestProduceActiveTunnelPackets() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	tunnel := types.Tunnel{
		ID:       1,
		FeePayer: feePayer.String(),
		IsActive: true,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x",
	}

	s.bandtssKeeper.EXPECT().IsReady(gomock.Any()).Return(true).AnyTimes()
	s.bandtssKeeper.EXPECT().GetSigningFee(gomock.Any()).Return(
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil,
	).Times(2)

	s.feedsKeeper.EXPECT().GetAllPrices(gomock.Any()).Return([]feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 50000, Timestamp: 1733000000},
	})

	spendableCoins := types.DefaultBasePacketFee.Add(sdk.NewCoin("uband", sdkmath.NewInt(20)))
	s.bankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).Return(spendableCoins)
	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), feePayer, types.ModuleName, types.DefaultBasePacketFee).
		Return(nil)

	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"chain-1",
		"0x",
		gomock.Any(),
		feePayer,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).Return(bandtsstypes.SigningID(1), nil)

	// set tunnel & latest price
	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	// set deposit to the tunnel to be able to activate
	tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)
	k.SetTunnel(ctx, tunnel)

	err = k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	k.SetLatestPrices(ctx, types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 6000000, Timestamp: 1732000000},
	}, 0))

	err = k.ProduceActiveTunnelPackets(ctx)
	s.Require().NoError(err)

	newTunnelInfo, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().True(newTunnelInfo.IsActive)
	s.Require().Equal(newTunnelInfo.Sequence, uint64(1))

	activeTunnels := k.GetActiveTunnelIDs(ctx)
	s.Require().Equal([]uint64{1}, activeTunnels)
}

func (s *KeeperTestSuite) TestProduceActiveTunnelPacketsNotEnoughMoney() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	tunnel := types.Tunnel{
		ID:       1,
		FeePayer: feePayer.String(),
		IsActive: true,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "0x",
		DestinationContractAddress: "0x",
	}

	s.bandtssKeeper.EXPECT().IsReady(gomock.Any()).Return(true).AnyTimes()
	s.bandtssKeeper.EXPECT().GetSigningFee(gomock.Any()).Return(
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil,
	)

	s.feedsKeeper.EXPECT().GetAllPrices(gomock.Any()).Return([]feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 50000, Timestamp: 1733000000},
	})
	s.bankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).
		Return(sdk.Coins{sdk.NewInt64Coin("uband", 1)})

	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	defaultParams := types.DefaultParams()
	err = k.SetParams(ctx, defaultParams)
	s.Require().NoError(err)

	// set deposit to the tunnel to be able to activate
	tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)
	k.SetTunnel(ctx, tunnel)

	err = k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	k.SetLatestPrices(ctx, types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PRICE_STATUS_AVAILABLE, SignalID: "CS:BAND-USD", Price: 600000, Timestamp: 1733000000},
	}, 0))

	err = k.ProduceActiveTunnelPackets(ctx)
	s.Require().NoError(err)

	newTunnelInfo, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().False(newTunnelInfo.IsActive)
	s.Require().Equal(newTunnelInfo.Sequence, uint64(0))

	activeTunnels := k.GetActiveTunnelIDs(ctx)
	s.Require().Len(activeTunnels, 0)
}
