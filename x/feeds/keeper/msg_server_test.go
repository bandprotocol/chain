package keeper_test

import "github.com/bandprotocol/chain/v2/x/feeds/types"

func (suite *KeeperTestSuite) TestMsgSubmitSignals() {
	testCases := []struct {
		name      string
		input     *types.MsgSubmitSignals
		expErr    bool
		expErrMsg string
		postCheck func()
	}{
		{
			name: "no delegation",
			input: &types.MsgSubmitSignals{
				Delegator: InvalidDelegator.String(),
				Signals: []types.Signal{
					{
						ID:    "crypto_price.bandusd",
						Power: 10,
					},
				},
			},
			expErr:    true,
			expErrMsg: "not enough delegation",
			postCheck: func() {},
		},
		{
			name: "1 signal more than delegations",
			input: &types.MsgSubmitSignals{
				Delegator: ValidDelegator.String(),
				Signals: []types.Signal{
					{
						ID:    "crypto_price.bandusd",
						Power: 1e10 + 1,
					},
				},
			},
			expErr:    true,
			expErrMsg: "not enough delegation",
			postCheck: func() {},
		},
		{
			name: "2 signals more than delegations",
			input: &types.MsgSubmitSignals{
				Delegator: ValidDelegator.String(),
				Signals: []types.Signal{
					{
						ID:    "crypto_price.bandusd",
						Power: 1e10,
					},
					{
						ID:    "crypto_price.atomusd",
						Power: 1,
					},
				},
			},
			expErr:    true,
			expErrMsg: "not enough delegation",
			postCheck: func() {},
		},
		{
			name: "valid request",
			input: &types.MsgSubmitSignals{
				Delegator: ValidDelegator.String(),
				Signals: []types.Signal{
					{
						ID:    "crypto_price.bandusd",
						Power: 1e10,
					},
				},
			},
			expErr:    false,
			expErrMsg: "",
			postCheck: func() {
				suite.Require().Equal(
					[]types.Signal{
						{
							ID:    "crypto_price.bandusd",
							Power: 1e10,
						},
					},
					suite.feedsKeeper.GetDelegatorSignals(suite.ctx, ValidDelegator),
				)
				suite.Require().Equal(
					[]types.Feed{
						{
							SignalID:                    "crypto_price.bandusd",
							Power:                       1e10,
							Interval:                    360,
							DeviationInThousandth:       30,
							LastIntervalUpdateTimestamp: suite.ctx.BlockTime().Unix(),
						},
					},
					suite.feedsKeeper.GetSupportedFeedsByPower(suite.ctx),
				)
			},
		},
		{
			name: "valid request (replace)",
			input: &types.MsgSubmitSignals{
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
			expErr:    false,
			expErrMsg: "",
			postCheck: func() {
				suite.Require().Equal(
					[]types.Signal{
						{
							ID:    "crypto_price.bandusd",
							Power: 1e9,
						},
						{
							ID:    "crypto_price.btcusd",
							Power: 1e9,
						},
					},
					suite.feedsKeeper.GetDelegatorSignals(suite.ctx, ValidDelegator),
				)
				suite.Require().Equal(
					[]types.Feed{
						{
							SignalID:                    "crypto_price.bandusd",
							Power:                       1e9,
							Interval:                    3600,
							DeviationInThousandth:       300,
							LastIntervalUpdateTimestamp: suite.ctx.BlockTime().Unix(),
						},
						{
							SignalID:                    "crypto_price.btcusd",
							Power:                       1e9,
							Interval:                    3600,
							DeviationInThousandth:       300,
							LastIntervalUpdateTimestamp: suite.ctx.BlockTime().Unix(),
						},
					},
					suite.feedsKeeper.GetSupportedFeedsByPower(suite.ctx),
				)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.SubmitSignals(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}

			tc.postCheck()
		})
	}
}

func (suite *KeeperTestSuite) TestMsgSubmitPrices() {
	suite.feedsKeeper.SetFeed(suite.ctx, types.Feed{
		SignalID:                    "crypto_price.bandusd",
		Power:                       10e6,
		Interval:                    100,
		LastIntervalUpdateTimestamp: suite.ctx.BlockTime().Unix(),
	})

	testCases := []struct {
		name      string
		input     *types.MsgSubmitPrices
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid validator",
			input: &types.MsgSubmitPrices{
				Validator: InvalidValidator.String(),
				Timestamp: suite.ctx.BlockTime().Unix(),
				Prices: []types.SubmitPrice{
					{
						PriceStatus: types.PriceStatusAvailable,
						SignalID:    "crypto_price.bandusd",
						Price:       10e12,
					},
				},
			},
			expErr:    true,
			expErrMsg: "not bonded validator",
		},
		{
			name: "invalid symbol",
			input: &types.MsgSubmitPrices{
				Validator: ValidValidator.String(),
				Timestamp: suite.ctx.BlockTime().Unix(),
				Prices: []types.SubmitPrice{
					{
						PriceStatus: types.PriceStatusAvailable,
						SignalID:    "crypto_price.btcusd",
						Price:       10e12,
					},
				},
			},
			expErr:    true,
			expErrMsg: "signal id is not supported",
		},
		{
			name: "invalid timestamp",
			input: &types.MsgSubmitPrices{
				Validator: ValidValidator.String(),
				Timestamp: suite.ctx.BlockTime().Unix() - 200,
				Prices: []types.SubmitPrice{
					{
						PriceStatus: types.PriceStatusAvailable,
						SignalID:    "crypto_price.bandusd",
						Price:       10e12,
					},
				},
			},
			expErr:    true,
			expErrMsg: "invalid timestamp",
		},
		{
			name: "valid message",
			input: &types.MsgSubmitPrices{
				Validator: ValidValidator.String(),
				Timestamp: suite.ctx.BlockTime().Unix(),
				Prices: []types.SubmitPrice{
					{
						PriceStatus: types.PriceStatusAvailable,
						SignalID:    "crypto_price.bandusd",
						Price:       10e12,
					},
				},
			},
			expErr:    false,
			expErrMsg: "",
		},
		{
			name: "price too fast",
			input: &types.MsgSubmitPrices{
				Validator: ValidValidator.String(),
				Timestamp: suite.ctx.BlockTime().Unix(),
				Prices: []types.SubmitPrice{
					{
						PriceStatus: types.PriceStatusAvailable,
						SignalID:    "crypto_price.bandusd",
						Price:       10e12,
					},
				},
			},
			expErr:    true,
			expErrMsg: "price is submitted too early",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.SubmitPrices(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgUpdatePriceService() {
	params := suite.feedsKeeper.GetParams(suite.ctx)
	priceService := types.DefaultPriceService()

	testCases := []struct {
		name      string
		input     *types.MsgUpdatePriceService
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid admin",
			input: &types.MsgUpdatePriceService{
				Admin:        "invalid",
				PriceService: priceService,
			},
			expErr:    true,
			expErrMsg: "invalid admin",
		},
		{
			name: "all good",
			input: &types.MsgUpdatePriceService{
				Admin:        params.Admin,
				PriceService: priceService,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdatePriceService(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	params := types.DefaultParams()

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: suite.feedsKeeper.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
