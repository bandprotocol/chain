package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGetSetLatestSignalPrices() {
	ctx, k := s.ctx, s.keeper

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
	s.Require().NoError(err)
	s.Require().Equal(latestSignalPrices, retrievedSignalPrices)
}

func (s *KeeperTestSuite) TestGetAllLatestSignalPrices() {
	ctx, k := s.ctx, s.keeper

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
	s.Require().Len(allLatestSignalPrices, 2)
	s.Require().Contains(allLatestSignalPrices, latestSignalPrices1)
	s.Require().Contains(allLatestSignalPrices, latestSignalPrices2)
}
