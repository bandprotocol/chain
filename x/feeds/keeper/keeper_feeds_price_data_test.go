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
		encoder      types.Encoder
		expectResult types.FeedsPriceData
		expectError  error
	}{
		{
			name:      "success case - fixed-point abi encoder",
			signalIDs: []string{"CS:atom-usd", "CS:band-usd"},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "CS:band-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			encoder: types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "CS:atom-usd",
						Price:    1e10,
					},
					{
						SignalID: "CS:band-usd",
						Price:    1e10,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
			expectError: nil,
		},
		{
			name:      "success case - tick abi encoder",
			signalIDs: []string{"CS:atom-usd", "CS:band-usd"},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "CS:band-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			encoder: types.ENCODER_TICK_ABI,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "CS:atom-usd",
						Price:    285171,
					},
					{
						SignalID: "CS:band-usd",
						Price:    285171,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
			expectError: nil,
		},
		{
			name:      "fail case - price not available",
			signalIDs: []string{"CS:atom-usd", "CS:band-usd"},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusUnavailable,
				},
			},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("CS:atom-usd: price not available"),
		},
		{
			name:      "fail case - price too old",
			signalIDs: []string{"CS:atom-usd"},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix() - 1000,
					PriceStatus: types.PriceStatusAvailable,
				},
			},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("CS:atom-usd: price too old"),
		},
		{
			name:         "fail case - price not found",
			signalIDs:    []string{"CS:atom-usdfake"},
			setPrices:    []types.Price{},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("failed to get price for signal id: CS:atom-usdfake: price not found"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Set up the prices
			suite.feedsKeeper.SetPrices(suite.ctx, tc.setPrices)
			feeds := make([]types.Feed, 0)
			for _, signalID := range tc.signalIDs {
				feeds = append(feeds, types.Feed{
					SignalID: signalID,
					Interval: 100,
				})
			}
			suite.feedsKeeper.SetCurrentFeeds(suite.ctx, feeds)

			// Call the function under test
			feedsPriceData, err := suite.feedsKeeper.GetFeedsPriceData(suite.ctx, tc.signalIDs, tc.encoder)

			// Check the result
			if tc.expectError != nil {
				suite.Require().ErrorContains(err, tc.expectError.Error())
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectResult, *feedsPriceData)
			}

			// cleanup
			// for _, price := range tc.setPrices {
			// 	suite.feedsKeeper.DeletePrice(suite.ctx, price.SignalID)
			// }
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
				"CS:atom-usd",
				"CS:band-usd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix(),
					PriceStatus: types.PriceStatusAvailable,
				},
				{
					SignalID:    "CS:band-usd",
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
				"CS:atom-usd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
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
				"CS:atom-usd",
			},
			setPrices: []types.Price{
				{
					SignalID:    "CS:atom-usd",
					Price:       1e10,
					Timestamp:   suite.ctx.BlockTime().Unix() - 1000,
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
			feeds := make([]types.Feed, 0)
			for _, signalID := range tc.signalIDs {
				feeds = append(feeds, types.Feed{
					SignalID: signalID,
					Interval: 100,
				})
			}
			suite.feedsKeeper.SetCurrentFeeds(suite.ctx, feeds)

			// Call the function under test
			_, err := suite.feedsKeeper.GetFeedsPriceData(
				suite.ctx,
				tc.signalIDs,
				types.ENCODER_FIXED_POINT_ABI,
			)

			// Check the result
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
