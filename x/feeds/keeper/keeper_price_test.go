package keeper_test

import (
	"errors"
	"time"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestGetSetDeletePrice() {
	ctx := suite.ctx

	// set
	expPrice := types.Price{
		Status:    types.PriceStatusAvailable,
		SignalID:  "CS:BAND-USD",
		Price:     1e10,
		Timestamp: ctx.BlockTime().Unix(),
	}
	suite.feedsKeeper.SetPrice(ctx, expPrice)

	// get
	price := suite.feedsKeeper.GetPrice(ctx, "CS:BAND-USD")
	suite.Require().Equal(expPrice, price)
}

func (suite *KeeperTestSuite) TestGetSetDeletePrices() {
	ctx := suite.ctx

	// set
	expPrices := []types.Price{
		{
			Status:    types.PriceStatusAvailable,
			SignalID:  "CS:ATOM-USD",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			Status:    types.PriceStatusAvailable,
			SignalID:  "CS:BAND-USD",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	suite.feedsKeeper.SetPrices(ctx, expPrices)

	// get all prices
	prices := suite.feedsKeeper.GetAllPrices(ctx)
	suite.Require().Equal(expPrices, prices)

	// get prices
	prices = suite.feedsKeeper.GetPrices(ctx, []string{"CS:ATOM-USD", "CS:BAND-USD"})
	suite.Require().Equal(expPrices, prices)

	// get prices not in store should return price with status not in current feeds
	expPrices = append(expPrices, types.Price{
		SignalID:  "CS:ETH-USD",
		Status:    types.PriceStatusNotInCurrentFeeds,
		Price:     0,
		Timestamp: 0,
	})
	prices = suite.feedsKeeper.GetPrices(ctx, []string{"CS:ATOM-USD", "CS:BAND-USD", "CS:ETH-USD"})
	suite.Require().Equal(expPrices, prices)

	// delete all
	suite.feedsKeeper.DeleteAllPrices(ctx)
	prices = suite.feedsKeeper.GetAllPrices(ctx)
	suite.Require().Empty(prices)
}

func (suite *KeeperTestSuite) TestGetSetValidatorPriceList() {
	ctx := suite.ctx

	// set
	expValPrices := []types.ValidatorPrice{
		{
			SignalPriceStatus: types.SignalPriceStatusAvailable,
			SignalID:          "CS:BAND-USD",
			Price:             1e10,
			Timestamp:         ctx.BlockTime().Unix(),
			BlockHeight:       ctx.BlockHeight(),
		},
		{
			SignalPriceStatus: types.SignalPriceStatusAvailable,
			SignalID:          "CS:ETH-USD",
			Price:             1e10 + 5,
			Timestamp:         ctx.BlockTime().Unix(),
			BlockHeight:       ctx.BlockHeight(),
		},
	}
	err := suite.feedsKeeper.SetValidatorPriceList(ctx, ValidValidator, expValPrices)
	suite.Require().NoError(err)

	// get
	valPrices, err := suite.feedsKeeper.GetValidatorPriceList(ctx, ValidValidator)
	suite.Require().NoError(err)
	suite.Require().Equal(expValPrices, valPrices.ValidatorPrices)
}

func (suite *KeeperTestSuite) TestCalculatePrices() {
	ctx := suite.ctx

	tests := []struct {
		name           string
		setup          func()
		expectError    bool
		expectedPrices []types.Price
	}{
		{
			name: "normal case with valid prices",
			setup: func() {
				// Set current feeds
				suite.feedsKeeper.SetCurrentFeeds(ctx, []types.Feed{
					{SignalID: "CS:BAND-USD", Interval: 60},
				})

				// Mock validators power
				suite.stakingKeeper.EXPECT().
					IterateBondedValidatorsByPower(ctx, gomock.Any()).
					DoAndReturn(func(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
						validators := []stakingtypes.Validator{
							{OperatorAddress: ValidValidator.String(), Tokens: sdkmath.NewInt(5000)},
							{OperatorAddress: ValidValidator2.String(), Tokens: sdkmath.NewInt(3000)},
						}

						for i, val := range validators {
							if stop := fn(int64(i), val); stop {
								break
							}
						}
						return nil
					})

				// Set validator prices
				err := suite.feedsKeeper.SetValidatorPriceList(ctx, ValidValidator, []types.ValidatorPrice{
					{
						SignalPriceStatus: types.SignalPriceStatusAvailable,
						SignalID:          "CS:BAND-USD",
						Price:             1000,
						Timestamp:         ctx.BlockTime().Unix(),
						BlockHeight:       ctx.BlockHeight(),
					},
				})
				suite.Require().NoError(err)

				err = suite.feedsKeeper.SetValidatorPriceList(ctx, ValidValidator2, []types.ValidatorPrice{
					{
						SignalID:          "CS:BAND-USD",
						SignalPriceStatus: types.SignalPriceStatusAvailable,
						Price:             2000,
						Timestamp:         ctx.BlockTime().Unix(),
						BlockHeight:       ctx.BlockHeight(),
					},
				})
				suite.Require().NoError(err)

				// Mock bonded tokens and quorum
				suite.stakingKeeper.EXPECT().TotalBondedTokens(ctx).Return(sdkmath.NewInt(11000), nil)
			},
			expectError: false,
			expectedPrices: []types.Price{
				{
					Status:    types.PriceStatusAvailable,
					SignalID:  "CS:BAND-USD",
					Price:     1000,
					Timestamp: ctx.BlockTime().Unix(),
				},
			},
		},
		{
			name: "error fetching total bonded tokens",
			setup: func() {
				// Set empty feeds
				suite.feedsKeeper.SetCurrentFeeds(ctx, []types.Feed{})

				// Mock validators power
				suite.stakingKeeper.EXPECT().
					IterateBondedValidatorsByPower(ctx, gomock.Any()).
					Return(nil)

				// Mock bonded tokens error
				suite.stakingKeeper.EXPECT().TotalBondedTokens(ctx).Return(sdkmath.ZeroInt(), errors.New("error"))
			},
			expectError: true,
		},
		{
			name: "error iterating validators",
			setup: func() {
				// Set empty feeds
				suite.feedsKeeper.SetCurrentFeeds(ctx, []types.Feed{})

				// Mock validators power error
				suite.stakingKeeper.EXPECT().
					IterateBondedValidatorsByPower(ctx, gomock.Any()).
					Return(errors.New("error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setup()
			err := suite.feedsKeeper.CalculatePrices(ctx)
			if tt.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				for _, expectedPrice := range tt.expectedPrices {
					price := suite.feedsKeeper.GetPrice(ctx, expectedPrice.SignalID)
					suite.Require().Equal(expectedPrice, price)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCalculatePrice() {
	ctx := suite.ctx

	// Define common variables
	feed := types.Feed{
		SignalID: "CS:BAND-USD",
		Interval: 60,
	}

	tests := []struct {
		name                string
		validatorPriceInfos []types.ValidatorPriceInfo
		powerQuorum         uint64
		expectedPrice       types.Price
		expectError         bool
	}{
		{
			name: "more than half have unsupported price status",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             1000,
					Price:             1000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusUnsupported,
					Power:             2001,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             1000,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
			},
			powerQuorum: 5000,
			expectedPrice: types.Price{
				Status:    types.PriceStatusUnknownSignalID,
				SignalID:  "CS:BAND-USD",
				Price:     0,
				Timestamp: ctx.BlockTime().Unix(),
			},
			expectError: false,
		},
		{
			name: "total power is less than quorum",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             1000,
					Price:             1000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             1000,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             1000,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
			},
			powerQuorum: 5000,
			expectedPrice: types.Price{
				Status:    types.PriceStatusNotReady,
				SignalID:  "CS:BAND-USD",
				Price:     0,
				Timestamp: ctx.BlockTime().Unix(),
			},
			expectError: false,
		},
		{
			name: "normal case",
			validatorPriceInfos: []types.ValidatorPriceInfo{
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             5000,
					Price:             1000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             3000,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
				{
					SignalPriceStatus: types.SignalPriceStatusAvailable,
					Power:             3000,
					Price:             2000,
					Timestamp:         ctx.BlockTime().Unix(),
				},
			},
			powerQuorum: 7000,
			expectedPrice: types.Price{
				Status:    types.PriceStatusAvailable,
				SignalID:  "CS:BAND-USD",
				Price:     1000,
				Timestamp: ctx.BlockTime().Unix(),
			},
			expectError: false,
		},
		{
			name:                "empty validator price infos",
			validatorPriceInfos: []types.ValidatorPriceInfo{},
			powerQuorum:         5000,
			expectedPrice: types.Price{
				Status:    types.PriceStatusNotReady,
				SignalID:  "CS:BAND-USD",
				Price:     0,
				Timestamp: ctx.BlockTime().Unix(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			price, err := suite.feedsKeeper.CalculatePrice(ctx, feed, tt.validatorPriceInfos, tt.powerQuorum)
			if tt.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expectedPrice, price)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCheckMissReport() {
	tests := []struct {
		name                string
		feed                types.Feed
		lastUpdateTimestamp int64
		lastUpdateBlock     int64
		valPrice            types.ValidatorPrice
		valInfo             types.ValidatorInfo
		blockTime           time.Time
		blockHeight         int64
		gracePeriod         int64
		expectedResult      bool
	}{
		{
			name: "Validator within grace period in time and blocks",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60, // Feed interval is 60 seconds
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1000,
				BlockHeight:       100,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0), // Validator activated at time 900
				},
			},
			blockTime:      time.Unix(1050, 0), // Current time is 1050
			blockHeight:    110,                // Current block height is 110
			gracePeriod:    120,                // Grace period is 120 seconds
			expectedResult: false,              // Should not get miss report
		},
		{
			name: "Validator outside grace period in time but within grace period in blocks",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     300,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1000,
				BlockHeight:       300,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1300, 0), // Current time is 1300
			blockHeight:    330,                // Current block height is 330
			gracePeriod:    120,
			expectedResult: false, // Still within grace period in blocks
		},
		{
			name: "Validator within grace period in time but outside in blocks",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1000,
				BlockHeight:       100,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1100, 0),
			blockHeight:    350, // Outside grace period in blocks
			gracePeriod:    120,
			expectedResult: false, // Still within grace period in time
		},
		{
			name: "Validator outside grace period and hasn't reported within feed interval",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1000,
				BlockHeight:       100,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1300, 0),
			blockHeight:    350, // Now outside grace period in blocks
			gracePeriod:    120,
			expectedResult: true, // Should get miss report
		},
		{
			name: "Validator outside grace period but has reported within feed interval",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1250, // Recent report
				BlockHeight:       330,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1300, 0),
			blockHeight:    350,
			gracePeriod:    120,
			expectedResult: false, // Should not get miss report
		},
		{
			name: "Validator outside grace period but reported just before feed interval expired",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1240,
				BlockHeight:       329,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1300, 0),
			blockHeight:    350,
			gracePeriod:    120,
			expectedResult: false, // Should not get miss report
		},
		{
			name: "Validator outside grace period and feed interval expired",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 1000,
			lastUpdateBlock:     100,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Timestamp:         1230,
				BlockHeight:       328,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(900, 0),
				},
			},
			blockTime:      time.Unix(1300, 0),
			blockHeight:    350,
			gracePeriod:    120,
			expectedResult: true, // Should get miss report
		},
		{
			name: "Validator has never reported but just activated",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 0,
			lastUpdateBlock:     0,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusUnspecified,
				Timestamp:         0,
				BlockHeight:       0,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(1000, 0),
				},
			},
			blockTime:      time.Unix(1120, 0),
			blockHeight:    350,
			gracePeriod:    120,
			expectedResult: false, // Still within grace period from activation
		},
		{
			name: "Validator never reported and outside grace period",
			feed: types.Feed{
				SignalID: "CS:BAND-USD",
				Interval: 60,
			},
			lastUpdateTimestamp: 0,
			lastUpdateBlock:     0,
			valPrice: types.ValidatorPrice{
				SignalPriceStatus: types.SignalPriceStatusUnspecified,
				Timestamp:         0,
				BlockHeight:       0,
			},
			valInfo: types.ValidatorInfo{
				Status: oracletypes.ValidatorStatus{
					IsActive: true,
					Since:    time.Unix(1000, 0),
				},
			},
			blockTime:      time.Unix(1300, 0),
			blockHeight:    500, // Far beyond grace period blocks
			gracePeriod:    120,
			expectedResult: true, // Should get miss report
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := keeper.CheckMissReport(
				tt.feed,
				tt.lastUpdateTimestamp,
				tt.lastUpdateBlock,
				tt.valPrice,
				tt.valInfo,
				tt.blockTime,
				tt.blockHeight,
				tt.gracePeriod,
			)
			suite.Require().Equal(tt.expectedResult, result)
		})
	}
}
