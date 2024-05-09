package keeper_test

import (
	gocontext "context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestQueryDelegatorSignals() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	signals := []types.Signal{
		{
			ID:    "crypto_price.bandusd",
			Power: 1e9,
		},
		{
			ID:    "crypto_price.btcusd",
			Power: 1e9,
		},
	}
	_, err := suite.msgServer.SubmitSignals(ctx, &types.MsgSubmitSignals{
		Delegator: ValidDelegator.String(),
		Signals:   signals,
	})
	suite.Require().NoError(err)

	// query and check
	res, err := queryClient.DelegatorSignals(gocontext.Background(), &types.QueryDelegatorSignalsRequest{
		Delegator: ValidDelegator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryDelegatorSignalsResponse{
		Signals: signals,
	}, res)
}

func (suite *KeeperTestSuite) TestQueryPrices() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	prices := []*types.Price{
		{
			SignalID:  "crypto_price.atomusd",
			Price:     100000000,
			Timestamp: 1234567890,
		},
		{
			SignalID:  "crypto_price.bandusd",
			Price:     200000000,
			Timestamp: 1234567890,
		},
	}

	for _, price := range prices {
		suite.feedsKeeper.SetPrice(ctx, *price)
	}

	// query and check
	var (
		req    *types.QueryPricesRequest
		expRes *types.QueryPricesResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"all prices",
			func() {
				req = &types.QueryPricesRequest{}
				expRes = &types.QueryPricesResponse{
					Prices: prices,
				}
			},
			true,
		},
		{
			"limit 1",
			func() {
				req = &types.QueryPricesRequest{
					Pagination: &query.PageRequest{Limit: 1},
				}
				expRes = &types.QueryPricesResponse{
					Prices: prices[:1],
				}
			},
			true,
		},
		{
			"filter",
			func() {
				req = &types.QueryPricesRequest{
					SignalIds: []string{"crypto_price.bandusd"},
				}
				expRes = &types.QueryPricesResponse{
					Prices: prices[1:],
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Prices(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetPrices(), res.GetPrices())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryPrice() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	feed := types.Feed{
		SignalID:                    "crypto_price.bandusd",
		Power:                       100000000,
		Interval:                    100,
		LastIntervalUpdateTimestamp: 1234567890,
	}
	suite.feedsKeeper.SetFeed(ctx, feed)

	price := types.Price{
		SignalID:  "crypto_price.bandusd",
		Price:     100000000,
		Timestamp: 1234567890,
	}
	suite.feedsKeeper.SetPrice(ctx, price)

	priceVal := types.ValidatorPrice{
		PriceStatus: types.PriceStatusAvailable,
		Validator:   ValidValidator.String(),
		SignalID:    "crypto_price.bandusd",
		Price:       1e9,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	err := suite.feedsKeeper.SetValidatorPrice(ctx, priceVal)
	suite.Require().NoError(err)

	// query and check
	res, err := queryClient.Price(gocontext.Background(), &types.QueryPriceRequest{
		SignalId: "crypto_price.bandusd",
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryPriceResponse{
		Price: price,
		ValidatorPrices: []types.ValidatorPrice{
			priceVal,
		},
	}, res)

	res, err = queryClient.Price(gocontext.Background(), &types.QueryPriceRequest{
		SignalId: "crypto_price.atomusd",
	})
	suite.Require().ErrorContains(err, "feed not found")
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryValidatorPrices() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	feeds := []types.Feed{
		{
			SignalID:                    "crypto_price.atomusd",
			Power:                       100000000,
			Interval:                    100,
			LastIntervalUpdateTimestamp: 1234567890,
		},
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       100000000,
			Interval:                    100,
			LastIntervalUpdateTimestamp: 1234567890,
		},
	}
	for _, feed := range feeds {
		suite.feedsKeeper.SetFeed(ctx, feed)
	}

	priceVals := []types.ValidatorPrice{
		{
			Validator: ValidValidator.String(),
			SignalID:  "crypto_price.atomusd",
			Price:     1e9,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			Validator: ValidValidator.String(),
			SignalID:  "crypto_price.bandusd",
			Price:     1e9,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	for _, priceVal := range priceVals {
		err := suite.feedsKeeper.SetValidatorPrice(ctx, priceVal)
		suite.Require().NoError(err)
	}

	// query and check
	res, err := queryClient.ValidatorPrices(gocontext.Background(), &types.QueryValidatorPricesRequest{
		Validator: ValidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: priceVals,
	}, res)

	res, err = queryClient.ValidatorPrices(gocontext.Background(), &types.QueryValidatorPricesRequest{
		Validator: InvalidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: nil,
	}, res)
}

func (suite *KeeperTestSuite) TestQueryValidatorPrice() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	priceVal := types.ValidatorPrice{
		Validator: ValidValidator.String(),
		SignalID:  "crypto_price.bandusd",
		Price:     1e9,
		Timestamp: ctx.BlockTime().Unix(),
	}
	err := suite.feedsKeeper.SetValidatorPrice(ctx, priceVal)
	suite.Require().NoError(err)

	// query and check
	res, err := queryClient.ValidatorPrice(gocontext.Background(), &types.QueryValidatorPriceRequest{
		SignalId:  "crypto_price.bandusd",
		Validator: ValidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPriceResponse{
		ValidatorPrice: priceVal,
	}, res)

	res, err = queryClient.ValidatorPrice(gocontext.Background(), &types.QueryValidatorPriceRequest{
		SignalId:  "crypto_price.atomusd",
		Validator: ValidValidator.String(),
	})
	suite.Require().ErrorContains(err, "validator price not found")
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryValidValidator() {
	queryClient := suite.queryClient

	// query and check
	res, err := queryClient.ValidValidator(gocontext.Background(), &types.QueryValidValidatorRequest{
		Validator: ValidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidValidatorResponse{
		Valid: true,
	}, res)

	res, err = queryClient.ValidValidator(gocontext.Background(), &types.QueryValidValidatorRequest{
		Validator: InvalidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidValidatorResponse{
		Valid: false,
	}, res)
}

func (suite *KeeperTestSuite) TestQueryFeeds() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	feeds := []*types.Feed{
		{
			SignalID:                    "crypto_price.atomusd",
			Power:                       100000000,
			Interval:                    100,
			LastIntervalUpdateTimestamp: 1234567890,
		},
		{
			SignalID:                    "crypto_price.bandusd",
			Power:                       100000000,
			Interval:                    100,
			LastIntervalUpdateTimestamp: 1234567890,
		},
	}

	for _, feed := range feeds {
		suite.feedsKeeper.SetFeed(ctx, *feed)
	}

	// query and check
	var (
		req    *types.QueryFeedsRequest
		expRes *types.QueryFeedsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"all feeds",
			func() {
				req = &types.QueryFeedsRequest{}
				expRes = &types.QueryFeedsResponse{
					Feeds: feeds,
				}
			},
			true,
		},
		{
			"limit 1",
			func() {
				req = &types.QueryFeedsRequest{
					Pagination: &query.PageRequest{Limit: 1},
				}
				expRes = &types.QueryFeedsResponse{
					Feeds: feeds[:1],
				}
			},
			true,
		},
		{
			"filter",
			func() {
				req = &types.QueryFeedsRequest{
					SignalIds: []string{"crypto_price.bandusd"},
				}
				expRes = &types.QueryFeedsResponse{
					Feeds: feeds[1:],
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.Feeds(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.GetFeeds(), res.GetFeeds())
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQuerySupportedFeeds() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	var (
		req    *types.QuerySupportedFeedsRequest
		expRes *types.QuerySupportedFeedsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no supported feeds",
			func() {
				req = &types.QuerySupportedFeedsRequest{}
				expRes = &types.QuerySupportedFeedsResponse{}
			},
			true,
		},
		{
			"1 supported symbol",
			func() {
				feeds := []types.Feed{
					{
						SignalID:                    "crypto_price.bandusd",
						Power:                       100000000,
						Interval:                    100,
						LastIntervalUpdateTimestamp: 1234567890,
					},
				}

				for _, feed := range feeds {
					suite.feedsKeeper.SetFeed(ctx, feed)
				}

				req = &types.QuerySupportedFeedsRequest{}
				expRes = &types.QuerySupportedFeedsResponse{
					Feeds: feeds,
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.SupportedFeeds(gocontext.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(expRes)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryPriceService() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	res, err := queryClient.PriceService(gocontext.Background(), &types.QueryPriceServiceRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryPriceServiceResponse{
		PriceService: suite.feedsKeeper.GetPriceService(ctx),
	}, res)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	res, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{
		Params: suite.feedsKeeper.GetParams(ctx),
	}, res)
}
