package keeper_test

import (
	"math"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
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
	s.Require().Equal(basePacketFee, totalFee.TotalPacketFee)
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

func (s *KeeperTestSuite) TestProducePacket() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	currentPricesMap := map[string]feedstypes.Price{
		"BTC/USD": {PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	}
	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	tunnel := types.Tunnel{
		ID:       1,
		FeePayer: feePayer.String(),
		IsActive: true,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(10)))).
		Return(nil)
	s.bandtssKeeper.EXPECT().GetParams(gomock.Any()).Return(bandtsstypes.Params{
		Fee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(10))),
	})
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"0x1234567890abcdef",
		"chain-1",
		gomock.Any(),
		feePayer,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(math.MaxInt))),
	).Return(bandtsstypes.SigningID(1), nil)

	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	// set deposit to the tunnel to be able to activate
	tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)

	k.SetTunnel(ctx, tunnel)

	err = k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	err = k.ProducePacket(ctx, tunnelID, currentPricesMap, false)
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
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x",
	}

	s.feedsKeeper.EXPECT().GetAllCurrentPrices(gomock.Any()).Return([]feedstypes.Price{
		{PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	})
	s.bankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).Return(types.DefaultBasePacketFee)
	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), feePayer, types.ModuleName, types.DefaultBasePacketFee).
		Return(nil)

	s.bandtssKeeper.EXPECT().GetParams(gomock.Any()).Return(bandtsstypes.Params{
		Fee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(10))),
	})
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"0x",
		"chain-1",
		gomock.Any(),
		feePayer,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(math.MaxInt))),
	).Return(bandtsstypes.SigningID(1), nil)

	// set tunnel & latest price
	err := tunnel.SetRoute(route)
	s.Require().NoError(err)

	// set deposit to the tunnel to be able to activate
	tunnel.TotalDeposit = append(tunnel.TotalDeposit, k.GetParams(ctx).MinDeposit...)
	k.SetTunnel(ctx, tunnel)

	err = k.ActivateTunnel(ctx, tunnelID)
	s.Require().NoError(err)

	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	// set params
	defaultParams := types.DefaultParams()
	err = k.SetParams(ctx, defaultParams)
	s.Require().NoError(err)

	k.ProduceActiveTunnelPackets(ctx)

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
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
		CreatedAt: ctx.BlockTime().Unix(),
	}
	route := &types.TSSRoute{
		DestinationChainID:         "0x",
		DestinationContractAddress: "0x",
	}

	s.feedsKeeper.EXPECT().GetAllCurrentPrices(gomock.Any()).Return([]feedstypes.Price{
		{PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
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

	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	k.ProduceActiveTunnelPackets(ctx)

	newTunnelInfo, err := k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().False(newTunnelInfo.IsActive)
	s.Require().Equal(newTunnelInfo.Sequence, uint64(0))

	activeTunnels := k.GetActiveTunnelIDs(ctx)
	s.Require().Len(activeTunnels, 0)
}

func (s *KeeperTestSuite) TestGenerateSignalPrices() {
	ctx := s.ctx

	tunnelID := uint64(1)
	currentPricesMap := map[string]feedstypes.Price{
		"BTC/USD": {PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	}
	triggerAll := true
	tunnel := types.Tunnel{
		ID: 1,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
	}
	latestSignalPrices := types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0)

	nsps := keeper.GenerateSignalPrices(
		ctx,
		currentPricesMap,
		tunnel.GetSignalDeviationMap(),
		latestSignalPrices.SignalPrices,
		triggerAll,
	)
	s.Require().Len(nsps, 1)
}
