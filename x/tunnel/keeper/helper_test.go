package keeper_test

import (
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGenerateSignalPricesSendAll() {
	ctx := s.ctx

	tunnelID := uint64(1)
	currentPricesMap := map[string]feedstypes.Price{
		"BTC/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	}
	sendAll := true
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
	}
	latestSignalPrices := types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 0},
	}, 0)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestSignalPrices(ctx, latestSignalPrices)

	newSignalPrices, err := keeper.GenerateNewSignalPrices(
		latestSignalPrices,
		tunnel.GetSignalDeviationMap(),
		currentPricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newSignalPrices, 1)
}

func (s *KeeperTestSuite) TestGenerateSignalPricesMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	currentPricesMap := map[string]feedstypes.Price{
		"BTC/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
		"ETH/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 2000, Timestamp: 0},
	}
	sendAll := false
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
			{SignalID: "ETH/USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
		},
	}
	latestSignalPrices := types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 48500}, // 3%
		{SignalID: "ETH/USD", Price: 1980},  // 1%
	}, 0)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestSignalPrices(ctx, latestSignalPrices)

	newSignalPrices, err := keeper.GenerateNewSignalPrices(
		latestSignalPrices,
		tunnel.GetSignalDeviationMap(),
		currentPricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newSignalPrices, 2)
}

func (s *KeeperTestSuite) TestGenerateSignalPricesNotMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	currentPricesMap := map[string]feedstypes.Price{
		"BTC/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
		"ETH/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 2000, Timestamp: 0},
	}
	sendAll := false
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
			{SignalID: "ETH/USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
		},
	}
	latestSignalPrices := types.NewLatestSignalPrices(tunnelID, []types.SignalPrice{
		{SignalID: "BTC/USD", Price: 49000}, // 2%
		{SignalID: "ETH/USD", Price: 1950},  // 2.5%
	}, 0)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestSignalPrices(ctx, latestSignalPrices)

	newSignalPrices, err := keeper.GenerateNewSignalPrices(
		latestSignalPrices,
		tunnel.GetSignalDeviationMap(),
		currentPricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newSignalPrices, 0)
}
