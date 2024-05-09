package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetDeleteSymbol() {
	ctx := suite.ctx

	// set
	expFeed := types.Feed{
		SignalID:                    "crypto_price.bandusd",
		Power:                       1e10,
		Interval:                    60,
		LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
	}
	suite.feedsKeeper.SetFeed(ctx, expFeed)

	// get
	feed, err := suite.feedsKeeper.GetFeed(ctx, "crypto_price.bandusd")
	suite.Require().NoError(err)
	suite.Require().Equal(expFeed, feed)

	// delete
	suite.feedsKeeper.DeleteFeed(ctx, feed)

	// get
	_, err = suite.feedsKeeper.GetFeed(ctx, "crypto_price.bandusd")
	suite.Require().ErrorContains(err, "feed not found")
}

func (suite *KeeperTestSuite) TestGetSetFeeds() {
	ctx := suite.ctx

	// set
	expFeeds := []types.Feed{
		{
			SignalID:                    "crypto_price.atomusd",
			Power:                       1e10,
			Interval:                    60,
			LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
		},
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       1e10,
			Interval:                    60,
			LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
		},
	}
	suite.feedsKeeper.SetFeeds(ctx, expFeeds)

	// get
	feeds := suite.feedsKeeper.GetFeeds(ctx)
	suite.Require().Equal(expFeeds, feeds)
}

func (suite *KeeperTestSuite) TestGetSetDeleteSymbolByPower() {
	ctx := suite.ctx

	// set
	expFeeds := []types.Feed{
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       1e10,
			Interval:                    60,
			LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
		},
		{
			SignalID:                    "crypto_price.atomusd",
			Power:                       1e9,
			Interval:                    60,
			LastIntervalUpdateTimestamp: ctx.BlockTime().Unix(),
		},
	}
	for _, expFeed := range expFeeds {
		suite.feedsKeeper.SetFeed(ctx, expFeed)
	}

	// get
	feeds := suite.feedsKeeper.GetSupportedFeedsByPower(ctx)
	suite.Require().Equal(expFeeds, feeds)

	// delete
	suite.feedsKeeper.DeleteFeed(ctx, expFeeds[0])

	// get
	feeds = suite.feedsKeeper.GetSupportedFeedsByPower(ctx)
	suite.Require().Equal(expFeeds[1:], feeds)
}
