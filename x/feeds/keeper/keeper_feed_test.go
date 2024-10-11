package keeper_test

import "github.com/bandprotocol/chain/v3/x/feeds/types"

func (suite *KeeperTestSuite) TestGetSetCurrentFeeds() {
	ctx := suite.ctx

	// set
	expFeed := []types.Feed{
		{
			SignalID: "CS:BAND-USD",
			Interval: 60,
		},
		{
			SignalID: "CS:ATOM-USD",
			Interval: 60,
		},
	}
	suite.feedsKeeper.SetCurrentFeeds(ctx, expFeed)

	// get
	feeds := suite.feedsKeeper.GetCurrentFeeds(ctx)
	suite.Require().Equal(expFeed, feeds.Feeds)
}

func (suite *KeeperTestSuite) TestCalculateNewCurrentFeeds() {
	ctx := suite.ctx

	suite.feedsKeeper.SetSignalTotalPower(ctx, types.Signal{
		ID:    "CS:BAND-USD",
		Power: 60000000000,
	})
	suite.feedsKeeper.SetSignalTotalPower(ctx, types.Signal{
		ID:    "CS:ATOM-USD",
		Power: 30000000000,
	})

	feeds := suite.feedsKeeper.CalculateNewCurrentFeeds(ctx)
	suite.Require().Equal([]types.Feed{
		{
			SignalID: "CS:BAND-USD",
			Power:    60000000000,
			Interval: 60,
		},
		{
			SignalID: "CS:ATOM-USD",
			Power:    30000000000,
			Interval: 120,
		},
	}, feeds)
}
