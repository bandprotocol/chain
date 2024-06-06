package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetFeedsPriceData() {
	// Define test cases
	testCases := []struct {
		name         string
		signalIDs    []string
		setPrices    []types.Price
		feedType     types.FeedType
		expectResult types.FeedsPriceData
		expectError  error
	}{
		{
			name:      "success case - default feed type",
			signalIDs: []string{"crypto_price.atomusd", "crypto_price.bandusd"},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "crypto_price.bandusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			feedType: types.FEED_TYPE_DEFAULT,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "crypto_price.atomusd",
						Price:    1e10,
					},
					{
						SignalID: "crypto_price.bandusd",
						Price:    1e10,
					},
				},
				Timestamp: suite.ctx.BlockTime().Unix(),
			},
			expectError: nil,
		},
		{
			name:      "success case - tick feed type",
			signalIDs: []string{"crypto_price.atomusd", "crypto_price.bandusd"},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "crypto_price.bandusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			feedType: types.FEED_TYPE_TICK,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "crypto_price.atomusd",
						Price:    285171,
					},
					{
						SignalID: "crypto_price.bandusd",
						Price:    285171,
					},
				},
				Timestamp: suite.ctx.BlockTime().Unix(),
			},
			expectError: nil,
		},
		{
			name:      "fail case - price not available",
			signalIDs: []string{"crypto_price.atomusd", "crypto_price.bandusd"},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusUnavailable,
				},
			},
			feedType:     types.FEED_TYPE_DEFAULT,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("crypto_price.atomusd: price not available"),
		},
		{
			name:      "fail case - price too old",
			signalIDs: []string{"crypto_price.atomusd"},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix() - int64(types.MAX_PRICE_TIME_DIFF.Seconds()) - 1,
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			feedType:     types.FEED_TYPE_DEFAULT,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("crypto_price.atomusd: price too old"),
		},
		{
			name:         "fail case - price not found",
			signalIDs:    []string{"crypto_price.atomusd"},
			setPrices:    []types.Price{},
			feedType:     types.FEED_TYPE_DEFAULT,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("failed to get price for signal id: crypto_price.atomusd: price not found"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Set up the prices
			suite.feedsKeeper.SetPrices(suite.ctx, tc.setPrices)

			// Call the function under test
			feedsPriceData, err := suite.feedsKeeper.GetFeedsPriceData(suite.ctx, tc.signalIDs, tc.feedType)

			// Check the result
			if tc.expectError != nil {
				suite.Require().ErrorContains(err, tc.expectError.Error())
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectResult, *feedsPriceData)
			}

			// cleanup
			for _, price := range tc.setPrices {
				suite.feedsKeeper.DeletePrice(suite.ctx, price.SignalID)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetFeedsPriceData2() {
	testCases := []struct {
		name        string
		signalIDs   []string
		setPrices   []types.Price
		expectError bool
	}{
		{
			name: "valid prices",
			signalIDs: []string{
				"crypto_price.atomusd",
				"crypto_price.bandusd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "crypto_price.bandusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			expectError: false,
		},
		{
			name: "price not available",
			signalIDs: []string{
				"crypto_price.atomusd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusUnavailable,
				},
			},
			expectError: true,
		},
		{
			name: "price too old",
			signalIDs: []string{
				"crypto_price.atomusd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "crypto_price.atomusd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix() - int64(types.MAX_PRICE_TIME_DIFF.Seconds()) - 1,
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Set up the prices
			suite.feedsKeeper.SetPrices(suite.ctx, tc.setPrices)

			// Call the function under test
			_, err := suite.feedsKeeper.GetFeedsPriceData(suite.ctx, tc.signalIDs, types.FEED_TYPE_DEFAULT)

			// Check the result
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
