package keeper_test

import (
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGetSetLatestPrices() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	latestPrices := types.LatestPrices{
		TunnelID: tunnelID,
		Prices: []feedstypes.Price{
			{
				Status:    feedstypes.PRICE_STATUS_AVAILABLE,
				SignalID:  "CS:BAND-USD",
				Price:     50000,
				Timestamp: 1733000000,
			},
		},
	}

	k.SetLatestPrices(ctx, latestPrices)

	retrievedPrices, err := k.GetLatestPrices(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(latestPrices, retrievedPrices)
}

func (s *KeeperTestSuite) TestGetAllLatestPrices() {
	ctx, k := s.ctx, s.keeper

	latestPrices1 := types.LatestPrices{
		TunnelID: 1,
		Prices: []feedstypes.Price{
			{
				Status:    feedstypes.PRICE_STATUS_AVAILABLE,
				SignalID:  "CS:BAND-USD",
				Price:     50000,
				Timestamp: 1733000000,
			},
		},
	}
	latestPrices2 := types.LatestPrices{
		TunnelID: 2,
		Prices: []feedstypes.Price{
			{
				Status:    feedstypes.PRICE_STATUS_AVAILABLE,
				SignalID:  "CS:ETH-USD",
				Price:     3000,
				Timestamp: 1733000000,
			},
		},
	}

	k.SetLatestPrices(ctx, latestPrices1)
	k.SetLatestPrices(ctx, latestPrices2)

	allLatestPrices := k.GetAllLatestPrices(ctx)
	s.Require().Len(allLatestPrices, 2)
	s.Require().Contains(allLatestPrices, latestPrices1)
	s.Require().Contains(allLatestPrices, latestPrices2)
}
