package keeper_test

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (suite *KeeperTestSuite) TestQueryVote() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	signals := []types.Signal{
		{
			ID:    "CS:BAND-USD",
			Power: 1e9,
		},
		{
			ID:    "CS:BTC-USD",
			Power: 1e9,
		},
	}
	_, err := suite.msgServer.Vote(ctx, &types.MsgVote{
		Voter:   ValidVoter.String(),
		Signals: signals,
	})
	suite.Require().NoError(err)

	// query and check
	res, err := queryClient.Vote(context.Background(), &types.QueryVoteRequest{
		Voter: ValidVoter.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryVoteResponse{
		Signals: signals,
	}, res)
}

func (suite *KeeperTestSuite) TestQueryPrice() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup

	price := types.Price{
		SignalID:  "CS:BAND-USD",
		Price:     100000000,
		Timestamp: 1234567890,
	}
	suite.feedsKeeper.SetPrice(ctx, price)

	// query and check
	res, err := queryClient.Price(context.Background(), &types.QueryPriceRequest{
		SignalId: "CS:BAND-USD",
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryPriceResponse{
		Price: price,
	}, res)

	res, err = queryClient.Price(context.Background(), &types.QueryPriceRequest{
		SignalId: "CS:ATOM-USD",
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryPriceResponse{
		Price: types.Price{
			Status:    types.PriceStatusNotInCurrentFeeds,
			SignalID:  "CS:ATOM-USD",
			Price:     0,
			Timestamp: 0,
		},
	}, res)
}

func (suite *KeeperTestSuite) TestQueryPrices() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// Setup multiple prices
	prices := []types.Price{
		{
			SignalID:  "CS:BAND-USD",
			Price:     100000000,
			Timestamp: 1234567890,
		},
		{
			SignalID:  "CS:ATOM-USD",
			Price:     200000000,
			Timestamp: 1234567890,
		},
		{
			SignalID:  "CS:BTC-USD",
			Price:     300000000,
			Timestamp: 1234567890,
		},
	}
	for _, price := range prices {
		suite.feedsKeeper.SetPrice(ctx, price)
	}

	testCases := []struct {
		name           string
		signalIDs      []string
		expectedPrices []types.Price
	}{
		{
			name:      "query multiple existing prices",
			signalIDs: []string{"CS:BAND-USD", "CS:BTC-USD"},
			expectedPrices: []types.Price{
				prices[0],
				prices[2],
			},
		},
		{
			name:      "query non-existing price",
			signalIDs: []string{"CS:NON-EXISTENT"},
			expectedPrices: []types.Price{
				{
					SignalID:  "CS:NON-EXISTENT",
					Status:    types.PriceStatusNotInCurrentFeeds,
					Price:     0,
					Timestamp: 0,
				},
			},
		},
		{
			name:      "query all existing prices",
			signalIDs: []string{"CS:BAND-USD", "CS:ATOM-USD", "CS:BTC-USD"},
			expectedPrices: []types.Price{
				prices[0],
				prices[1],
				prices[2],
			},
		},
		{
			name:      "query with existing prices and non-existing price",
			signalIDs: []string{"CS:BAND-USD", "CS:NON-EXISTENT"},
			expectedPrices: []types.Price{
				prices[0],
				{SignalID: "CS:NON-EXISTENT", Status: types.PriceStatusNotInCurrentFeeds, Price: 0, Timestamp: 0},
			},
		},
		{
			name:           "query with empty signal IDs",
			signalIDs:      []string{},
			expectedPrices: []types.Price(nil),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Query prices
			res, err := queryClient.Prices(context.Background(), &types.QueryPricesRequest{
				SignalIds: tc.signalIDs,
			})
			suite.Require().NoError(err)
			suite.Require().Equal(&types.QueryPricesResponse{
				Prices: tc.expectedPrices,
			}, res)
		})
	}
}

func (suite *KeeperTestSuite) TestQueryAllPrices() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// Setup multiple prices
	prices := []types.Price{
		{
			SignalID:  "CS:ATOM-USD",
			Price:     200000000,
			Timestamp: 1234567891,
		},
		{
			SignalID:  "CS:BAND-USD",
			Price:     100000000,
			Timestamp: 1234567890,
		},
		{
			SignalID:  "CS:BTC-USD",
			Price:     300000000,
			Timestamp: 1234567892,
		},
	}
	for _, price := range prices {
		suite.feedsKeeper.SetPrice(ctx, price)
	}

	// Query all prices without pagination
	res, err := queryClient.AllPrices(context.Background(), &types.QueryAllPricesRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(len(prices), len(res.Prices))
	suite.Require().Equal(prices, res.Prices)

	// Query all prices with pagination
	pageReq := &query.PageRequest{Limit: 2, CountTotal: true}
	res, err = queryClient.AllPrices(context.Background(), &types.QueryAllPricesRequest{
		Pagination: pageReq,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(2, len(res.Prices))
	suite.Require().Equal(uint64(3), res.Pagination.Total)
	suite.Require().NotNil(res.Pagination.NextKey)

	// Query the next page
	resNext, err := queryClient.AllPrices(context.Background(), &types.QueryAllPricesRequest{
		Pagination: &query.PageRequest{Key: res.Pagination.NextKey, Limit: 2},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(resNext.Prices))
	suite.Require().Nil(resNext.Pagination.NextKey)
}

func (suite *KeeperTestSuite) TestQueryValidatorPrices() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	feeds := []types.Feed{
		{
			SignalID: "CS:ATOM-USD",
			Interval: 100,
		},
		{
			SignalID: "CS:BAND-USD",
			Interval: 100,
		},
	}

	suite.feedsKeeper.SetCurrentFeeds(ctx, feeds)

	valPrices := []types.ValidatorPrice{
		{
			SignalID:  "CS:ATOM-USD",
			Price:     1e9,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			SignalID:  "CS:BAND-USD",
			Price:     1e9,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}

	err := suite.feedsKeeper.SetValidatorPriceList(ctx, ValidValidator, valPrices)
	suite.Require().NoError(err)

	// query all prices
	res, err := queryClient.ValidatorPrices(context.Background(), &types.QueryValidatorPricesRequest{
		Validator: ValidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: valPrices,
	}, res)

	// query with specific SignalIds
	res, err = queryClient.ValidatorPrices(context.Background(), &types.QueryValidatorPricesRequest{
		Validator: ValidValidator.String(),
		SignalIds: []string{"CS:ATOM-USD"},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: []types.ValidatorPrice(nil),
	}, res)

	// query with invalid validator
	res, err = queryClient.ValidatorPrices(context.Background(), &types.QueryValidatorPricesRequest{
		Validator: InvalidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: []types.ValidatorPrice(nil),
	}, res)

	// query with specific SignalIds for invalid validator
	res, err = queryClient.ValidatorPrices(context.Background(), &types.QueryValidatorPricesRequest{
		Validator: InvalidValidator.String(),
		SignalIds: []string{"CS:ATOM-USD"},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidatorPricesResponse{
		ValidatorPrices: []types.ValidatorPrice(nil),
	}, res)
}

func (suite *KeeperTestSuite) TestQueryValidValidator() {
	queryClient := suite.queryClient

	// query and check
	res, err := queryClient.ValidValidator(context.Background(), &types.QueryValidValidatorRequest{
		Validator: ValidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidValidatorResponse{
		Valid: true,
	}, res)

	res, err = queryClient.ValidValidator(context.Background(), &types.QueryValidValidatorRequest{
		Validator: InvalidValidator.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryValidValidatorResponse{
		Valid: false,
	}, res)
}

func (suite *KeeperTestSuite) TestQuerySignalTotalPowers() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// setup
	signals := []*types.Signal{
		{
			ID:    "CS:ATOM-USD",
			Power: 100000000,
		},
		{
			ID:    "CS:BAND-USD",
			Power: 100000000,
		},
	}

	for _, signal := range signals {
		suite.feedsKeeper.SetSignalTotalPower(ctx, *signal)
	}

	// query and check
	var (
		req    *types.QuerySignalTotalPowersRequest
		expRes *types.QuerySignalTotalPowersResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"all feeds",
			func() {
				req = &types.QuerySignalTotalPowersRequest{}
				expRes = &types.QuerySignalTotalPowersResponse{
					SignalTotalPowers: signals,
				}
			},
			true,
		},
		{
			"limit 1",
			func() {
				req = &types.QuerySignalTotalPowersRequest{
					Pagination: &query.PageRequest{Limit: 1},
				}
				expRes = &types.QuerySignalTotalPowersResponse{
					SignalTotalPowers: signals[:1],
				}
			},
			true,
		},
		{
			"filter",
			func() {
				req = &types.QuerySignalTotalPowersRequest{
					SignalIds: []string{"CS:BAND-USD"},
				}
				expRes = &types.QuerySignalTotalPowersResponse{
					SignalTotalPowers: signals[1:],
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.SignalTotalPowers(context.Background(), req)

			if testCase.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.SignalTotalPowers, res.SignalTotalPowers)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCurrentFeeds() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	var (
		req    *types.QueryCurrentFeedsRequest
		expRes *types.QueryCurrentFeedsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no current feeds",
			func() {
				req = &types.QueryCurrentFeedsRequest{}
				expRes = &types.QueryCurrentFeedsResponse{
					CurrentFeeds: types.CurrentFeedWithDeviations{
						Feeds:               nil,
						LastUpdateTimestamp: ctx.BlockTime().Unix(),
						LastUpdateBlock:     ctx.BlockHeight(),
					},
				}
			},
			true,
		},
		{
			"1 current symbol",
			func() {
				feeds := []types.Feed{
					{
						SignalID: "CS:BAND-USD",
						Power:    36000000000,
						Interval: 100,
					},
				}

				suite.feedsKeeper.SetCurrentFeeds(ctx, feeds)

				feedWithDeviations := []types.FeedWithDeviation{
					{
						SignalID:            "CS:BAND-USD",
						Power:               36000000000,
						Interval:            100,
						DeviationBasisPoint: 83,
					},
				}

				req = &types.QueryCurrentFeedsRequest{}
				expRes = &types.QueryCurrentFeedsResponse{
					CurrentFeeds: types.CurrentFeedWithDeviations{
						Feeds:               feedWithDeviations,
						LastUpdateTimestamp: ctx.BlockTime().Unix(),
						LastUpdateBlock:     ctx.BlockHeight(),
					},
				}
			},
			true,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
			testCase.malleate()

			res, err := queryClient.CurrentFeeds(context.Background(), req)

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

func (suite *KeeperTestSuite) TestQueryReferenceSourceConfig() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	res, err := queryClient.ReferenceSourceConfig(context.Background(), &types.QueryReferenceSourceConfigRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryReferenceSourceConfigResponse{
		ReferenceSourceConfig: suite.feedsKeeper.GetReferenceSourceConfig(ctx),
	}, res)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx, queryClient := suite.ctx, suite.queryClient

	// query and check
	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(&types.QueryParamsResponse{
		Params: suite.feedsKeeper.GetParams(ctx),
	}, res)
}
