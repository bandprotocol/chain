package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestDeductBasePacketFee(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
	feePayer := sdk.AccAddress([]byte("fee_payer_address"))
	basePacketFee := sdk.Coins{sdk.NewInt64Coin("uband", 100)}

	defaultParams := types.DefaultParams()
	defaultParams.BasePacketFee = basePacketFee

	err := k.SetParams(ctx, defaultParams)
	require.NoError(t, err)

	// Mock bankKeeper to simulate coin transfer
	s.MockBankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, basePacketFee).
		Return(nil)

	// Call the DeductBasePacketFee function
	err = k.DeductBasePacketFee(ctx, feePayer)
	require.NoError(t, err)

	// Validate the total fees are updated
	totalFee := k.GetTotalFees(ctx)
	require.Equal(t, basePacketFee, totalFee.TotalPacketFee)
}

func TestGetSetPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx := s.Ctx
	k := s.Keeper

	packet := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}

	k.SetPacket(ctx, packet)

	storedPacket, err := k.GetPacket(ctx, packet.TunnelID, packet.Nonce)
	require.NoError(t, err)
	require.Equal(t, packet, storedPacket)
}

func TestProducePacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
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
		DestinationChainID:         "0x",
		DestinationContractAddress: "0x",
	}
	err := tunnel.SetRoute(route)
	require.NoError(t, err)

	// Mock bankKeeper to simulate coin transfer
	s.MockBankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, nil).
		Return(nil)

	// Set the tunnel
	k.SetTunnel(ctx, tunnel)
	err = k.ActivateTunnel(ctx, tunnelID)
	require.NoError(t, err)
	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	// Call the ProduceActiveTunnelPackets function
	err = k.ProducePacket(ctx, tunnelID, currentPricesMap, false)
	require.NoError(t, err)
}

func TestProduceActiveTunnelPackets(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
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
	err := tunnel.SetRoute(route)
	require.NoError(t, err)

	// set params
	defaultParams := types.DefaultParams()
	err = k.SetParams(ctx, defaultParams)
	require.NoError(t, err)

	// Set the tunnel
	k.SetTunnel(ctx, tunnel)
	err = k.ActivateTunnel(ctx, tunnelID)
	require.NoError(t, err)
	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	// Mock bankKeeper & FeedsKeeper
	s.MockFeedsKeeper.EXPECT().GetCurrentPrices(gomock.Any()).Return([]feedstypes.Price{
		{PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	})
	s.MockBankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).Return(types.DefaultBasePacketFee)
	s.MockBankKeeper.EXPECT().
		SendCoinsFromAccountToModule(gomock.Any(), feePayer, types.ModuleName, types.DefaultBasePacketFee).
		Return(nil)

	// Call the ProduceActiveTunnelPackets function
	k.ProduceActiveTunnelPackets(ctx)

	// Validate the tunnel is Inactive & no packet is Created
	newTunnelInfo, err := k.GetTunnel(ctx, tunnelID)
	require.NoError(t, err)
	require.True(t, newTunnelInfo.IsActive)
	require.Equal(t, newTunnelInfo.NonceCount, uint64(1))

	activeTunnels := k.GetActiveTunnelIDs(ctx)
	require.Equal(t, []uint64{1}, activeTunnels)
}

func TestProduceActiveTunnelPacketsNotEnoughMoney(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define test data
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
	err := tunnel.SetRoute(route)
	require.NoError(t, err)

	// set params
	defaultParams := types.DefaultParams()
	err = k.SetParams(ctx, defaultParams)
	require.NoError(t, err)

	// Set the tunnel
	k.SetTunnel(ctx, tunnel)
	err = k.ActivateTunnel(ctx, tunnelID)
	require.NoError(t, err)
	k.SetLatestSignalPrices(ctx, types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0))

	// Mock bankKeeper & FeedsKeeper
	s.MockFeedsKeeper.EXPECT().GetCurrentPrices(gomock.Any()).Return([]feedstypes.Price{
		{PriceStatus: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	})
	s.MockBankKeeper.EXPECT().SpendableCoins(gomock.Any(), feePayer).
		Return(sdk.Coins{sdk.NewInt64Coin("uband", 1)})

	// Call the ProduceActiveTunnelPackets function
	k.ProduceActiveTunnelPackets(ctx)

	// Validate the tunnel is Inactive & no packet is Created
	newTunnelInfo, err := k.GetTunnel(ctx, tunnelID)
	require.NoError(t, err)
	require.False(t, newTunnelInfo.IsActive)
	require.Equal(t, newTunnelInfo.NonceCount, uint64(0))

	activeTunnels := k.GetActiveTunnelIDs(ctx)
	require.Len(t, activeTunnels, 0)
}

func TestGenerateSignalPrices(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx := s.Ctx

	// Define test data
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

	// Call the GenerateSignalPrices function
	nsps := keeper.GenerateSignalPrices(
		ctx,
		currentPricesMap,
		tunnel.GetSignalDeviationMap(),
		latestSignalPrices.SignalPrices,
		triggerAll,
	)
	require.Len(t, nsps, 1)
}
