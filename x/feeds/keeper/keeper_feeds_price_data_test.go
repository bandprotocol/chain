package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
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
			signalIDs: []string{"CS:ATOM-USD", "CS:BAND-USD"},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
			},
			encoder: types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "CS:ATOM-USD",
						Price:    1e10,
					},
					{
						SignalID: "CS:BAND-USD",
						Price:    1e10,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
			expectError: nil,
		},
		{
			name:      "success case - tick abi encoder",
			signalIDs: []string{"CS:ATOM-USD", "CS:BAND-USD"},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
			},
			encoder: types.ENCODER_TICK_ABI,
			expectResult: types.FeedsPriceData{
				SignalPrices: []types.SignalPrice{
					{
						SignalID: "CS:ATOM-USD",
						Price:    285171,
					},
					{
						SignalID: "CS:BAND-USD",
						Price:    285171,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
			expectError: nil,
		},
		{
			name:      "fail case - price not in current feeds",
			signalIDs: []string{"CS:ATOM-USD", "CS:BAND-USD"},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusNotInCurrentFeeds,
				},
			},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("CS:ATOM-USD: price not available"),
		},
		{
			name:      "fail case - price too old",
			signalIDs: []string{"CS:ATOM-USD"},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix() - 1000,
					Status:    types.PriceStatusAvailable,
				},
			},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("CS:ATOM-USD: price too old"),
		},
		{
			name:         "fail case - price not found",
			signalIDs:    []string{"CS:ATOM-USDfake"},
			setPrices:    []types.Price{},
			encoder:      types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{},
			expectError:  fmt.Errorf("CS:ATOM-USDfake: price not available"),
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
				"CS:ATOM-USD",
				"CS:BAND-USD",
			},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
			},
			expectError: false,
		},
		{
			name: "price not in current feeds",
			signalIDs: []string{
				"CS:ATOM-USD",
			},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusNotInCurrentFeeds,
				},
			},
			expectError: true,
		},
		{
			name: "price too old",
			signalIDs: []string{
				"CS:ATOM-USD",
			},
			setPrices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: suite.ctx.BlockTime().Unix() - 1000,
					Status:    types.PriceStatusAvailable,
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
