package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.ctx

	err := suite.feedsKeeper.SetParams(ctx, types.DefaultParams())
	suite.Require().NoError(err)

	err = suite.feedsKeeper.SetPriceService(ctx, types.DefaultPriceService())
	suite.Require().NoError(err)

	feeds := []types.Feed{
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       10000,
			Interval:                    60,
			LastIntervalUpdateTimestamp: 123456789,
		},
	}
	suite.feedsKeeper.SetFeeds(ctx, feeds)

	exportGenesis := suite.feedsKeeper.ExportGenesis(ctx)

	suite.Require().Equal(types.DefaultParams(), exportGenesis.Params)
	suite.Require().Equal(types.DefaultPriceService(), exportGenesis.PriceService)
	suite.Require().Equal(feeds, exportGenesis.Feeds)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	feeds := []types.Feed{
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       10000,
			Interval:                    60,
			LastIntervalUpdateTimestamp: 123456789,
		},
	}

	delegatorSignals := []types.DelegatorSignals{
		{
			Delegator: ValidDelegator.String(),
			Signals: []types.Signal{
				{
					ID:    "crypto_price.bandusd",
					Power: 1e9,
				},
				{
					ID:    "crypto_price.btcusd",
					Power: 1e9,
				},
			},
		},
	}

	g := types.DefaultGenesisState()
	g.Feeds = feeds
	g.DelegatorSignals = delegatorSignals

	suite.feedsKeeper.InitGenesis(suite.ctx, *g)

	suite.Require().Equal(feeds, suite.feedsKeeper.GetFeeds(suite.ctx))
	suite.Require().Equal(types.DefaultPriceService(), suite.feedsKeeper.GetPriceService(ctx))
	suite.Require().Equal(types.DefaultParams(), suite.feedsKeeper.GetParams(ctx))
	for _, ds := range delegatorSignals {
		suite.Require().
			Equal(ds.Signals, suite.feedsKeeper.GetDelegatorSignals(ctx, sdk.MustAccAddressFromBech32(ds.Delegator)))
	}
}
