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

	delegatorSignals := []types.DelegatorSignals{
		{
			Delegator: ValidDelegator.String(),
			Signals: []types.Signal{
				{
					ID:    "crypto_price.bandusd",
					Power: 10000 * 1e6,
				},
				{
					ID:    "crypto_price.btcusd",
					Power: 20000 * 1e9,
				},
			},
		},
		{
			Delegator: ValidDelegator2.String(),
			Signals: []types.Signal{
				{
					ID:    "crypto_price.bandusd",
					Power: 20000 * 1e6,
				},
				{
					ID:    "crypto_price.btcusd",
					Power: 40000 * 1e9,
				},
			},
		},
	}
	suite.feedsKeeper.SetAllDelegatorSignals(ctx, delegatorSignals)

	exportGenesis := suite.feedsKeeper.ExportGenesis(ctx)

	suite.Require().Equal(types.DefaultParams(), exportGenesis.Params)
	suite.Require().Equal(types.DefaultPriceService(), exportGenesis.PriceService)
	suite.Require().Equal(delegatorSignals, exportGenesis.DelegatorSignals)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx
	params := types.NewParams("[NOT_SET]", 30, 30, 60, 3600, 1000_000_000, 100, 30, 5, 300, 256, 28800)

	delegatorSignals := []types.DelegatorSignals{
		{
			Delegator: ValidDelegator.String(),
			Signals: []types.Signal{
				{
					ID:    "crypto_price.bandusd",
					Power: 10000 * 1e6,
				},
				{
					ID:    "crypto_price.btcusd",
					Power: 20000 * 1e6,
				},
			},
		},
		{
			Delegator: ValidDelegator2.String(),
			Signals: []types.Signal{
				{
					ID:    "crypto_price.bandusd",
					Power: 20000 * 1e6,
				},
				{
					ID:    "crypto_price.btcusd",
					Power: 40000 * 1e6,
				},
			},
		},
	}

	g := types.DefaultGenesisState()
	g.DelegatorSignals = delegatorSignals
	g.Params = params

	suite.feedsKeeper.InitGenesis(suite.ctx, *g)

	suite.Require().Equal(types.DefaultPriceService(), suite.feedsKeeper.GetPriceService(ctx))
	suite.Require().Equal(params, suite.feedsKeeper.GetParams(ctx))
	for _, ds := range delegatorSignals {
		suite.Require().
			Equal(ds.Signals, suite.feedsKeeper.GetDelegatorSignals(ctx, sdk.MustAccAddressFromBech32(ds.Delegator)))
	}

	stpBand, err := suite.feedsKeeper.GetSignalTotalPower(ctx, "crypto_price.bandusd")
	suite.Require().NoError(err)
	suite.Require().Equal(types.Signal{
		ID:    "crypto_price.bandusd",
		Power: 30000 * 1e6,
	}, stpBand)

	stpBtc, err := suite.feedsKeeper.GetSignalTotalPower(ctx, "crypto_price.btcusd")
	suite.Require().NoError(err)
	suite.Require().Equal(types.Signal{
		ID:    "crypto_price.btcusd",
		Power: 60000 * 1e6,
	}, stpBtc)

	suite.Require().Equal(types.SupportedFeeds{
		Feeds: []types.Feed{
			{
				SignalID:              "crypto_price.btcusd",
				Interval:              60,
				DeviationInThousandth: 5,
			},
			{
				SignalID:              "crypto_price.bandusd",
				Interval:              120,
				DeviationInThousandth: 10,
			},
		},
		LastUpdateTimestamp: ctx.BlockTime().Unix(),
		LastUpdateBlock:     ctx.BlockHeight(),
	}, suite.feedsKeeper.GetSupportedFeeds(ctx))
}
