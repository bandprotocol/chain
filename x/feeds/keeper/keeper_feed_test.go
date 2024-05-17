package keeper_test

import "github.com/bandprotocol/chain/v2/x/feeds/types"

func (suite *KeeperTestSuite) TestGetSetSupportedFeeds() {
	ctx := suite.ctx

	// set
	expFeed := []types.Feed{
		{
			SignalID:              "crypto_price.bandusd",
			Interval:              60,
			DeviationInThousandth: 1000,
		},
		{
			SignalID:              "crypto_price.atomusd",
			Interval:              60,
			DeviationInThousandth: 1000,
		},
	}
	suite.feedsKeeper.SetSupportedFeeds(ctx, expFeed)

	// get
	feeds := suite.feedsKeeper.GetSupportedFeeds(ctx)
	suite.Require().Equal(expFeed, feeds.Feeds)
}

func (suite *KeeperTestSuite) TestCalculateNewSupportedFeeds() {
	ctx := suite.ctx

	suite.feedsKeeper.SetSignalTotalPower(ctx, types.Signal{
		ID:    "crypto_price.bandusd",
		Power: 60000000000,
	})
	suite.feedsKeeper.SetSignalTotalPower(ctx, types.Signal{
		ID:    "crypto_price.atomusd",
		Power: 30000000000,
	})

	feeds := suite.feedsKeeper.CalculateNewSupportedFeeds(ctx)
	suite.Require().Equal([]types.Feed{
		{
			SignalID:              "crypto_price.bandusd",
			Interval:              60,
			DeviationInThousandth: 5,
		},
		{
			SignalID:              "crypto_price.atomusd",
			Interval:              120,
			DeviationInThousandth: 10,
		},
	}, feeds)
}
