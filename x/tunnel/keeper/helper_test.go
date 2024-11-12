package keeper_test

import (
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGeneratePricesSendAll() {
	ctx := s.ctx

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
		"BTC/USD": {Status: feedstypes.PriceStatusAvailable, SignalID: "BTC/USD", Price: 50000, Timestamp: 0},
	}
	sendAll := true
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "BTC/USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
	}
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PriceStatusNotInCurrentFeeds, SignalID: "BTC/USD", Price: 0},
	}, 0)

	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices, err := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newPrices, 1)
}

func (s *KeeperTestSuite) TestGeneratePricesMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
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
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PriceStatusNotInCurrentFeeds, SignalID: "BTC/USD", Price: 48500}, // 3%
		{Status: feedstypes.PriceStatusNotInCurrentFeeds, SignalID: "ETH/USD", Price: 1980},  // 1%
	}, 0)
	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices, err := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newPrices, 2)
}

func (s *KeeperTestSuite) TestGeneratePricesNotMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
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
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{Status: feedstypes.PriceStatusNotInCurrentFeeds, SignalID: "BTC/USD", Price: 49000}, // 2%
		{Status: feedstypes.PriceStatusNotInCurrentFeeds, SignalID: "ETH/USD", Price: 1950},  // 2.5%
	}, 0)
	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices, err := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		sendAll,
	)
	s.Require().NoError(err)
	s.Require().Len(newPrices, 0)
}
