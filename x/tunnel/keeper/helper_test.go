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
		"CS:BAND-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     50000,
			Timestamp: 1733000000,
		},
	}
	sendAll := true
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 1000, HardDeviationBPS: 1000},
		},
	}
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{
			Status:    feedstypes.PRICE_STATUS_NOT_IN_CURRENT_FEEDS,
			SignalID:  "CS:BAND-USD",
			Price:     0,
			Timestamp: 1733000000,
		},
	}, 0)

	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		1733000000,
		sendAll,
	)
	s.Require().Len(newPrices, 1)
}

func (s *KeeperTestSuite) TestGeneratePricesMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
		"CS:BAND-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     50000,
			Timestamp: 1733000000,
		},
		"CS:ETH-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:ETH-USD",
			Price:     2000,
			Timestamp: 1733000000,
		},
	}
	sendAll := false
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
			{SignalID: "CS:ETH-USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
		},
	}
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     48500,
			Timestamp: 1732000000,
		}, // 3%
		{
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:ETH-USD",
			Price:     1980,
			Timestamp: 1732000000,
		}, // 1%
	}, 0)
	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		1733000000,
		sendAll,
	)
	s.Require().Len(newPrices, 2)
}

func (s *KeeperTestSuite) TestGeneratePricesNotMeetHardDeviation() {
	ctx := s.ctx

	tunnelID := uint64(1)
	pricesMap := map[string]feedstypes.Price{
		"CS:BAND-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     50000,
			Timestamp: 1733000000,
		},
		"CS:ETH-USD": {
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:ETH-USD",
			Price:     2000,
			Timestamp: 1733000000,
		},
	}
	sendAll := false
	tunnel := types.Tunnel{
		ID: tunnelID,
		SignalDeviations: []types.SignalDeviation{
			{SignalID: "CS:BAND-USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
			{SignalID: "CS:ETH-USD", SoftDeviationBPS: 100, HardDeviationBPS: 300},
		},
	}
	latestPrices := types.NewLatestPrices(tunnelID, []feedstypes.Price{
		{
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:BAND-USD",
			Price:     49000,
			Timestamp: 1732000000,
		}, // 2%
		{
			Status:    feedstypes.PRICE_STATUS_AVAILABLE,
			SignalID:  "CS:ETH-USD",
			Price:     1950,
			Timestamp: 1732000000,
		}, // 2.5%
	}, 0)
	latestPricesMap := keeper.CreatePricesMap(latestPrices.Prices)

	s.keeper.SetTunnel(ctx, tunnel)
	s.keeper.SetLatestPrices(ctx, latestPrices)

	newPrices := keeper.GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		pricesMap,
		1733000000,
		sendAll,
	)
	s.Require().Len(newPrices, 0)
}
