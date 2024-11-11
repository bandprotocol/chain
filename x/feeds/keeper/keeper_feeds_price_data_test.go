package keeper_test

import (
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
				Prices: []types.Price{
					{
						Timestamp: suite.ctx.BlockTime().Unix(),
						Status:    types.PriceStatusAvailable,
						SignalID:  "CS:ATOM-USD",
						Price:     1e10,
					},
					{
						SignalID:  "CS:BAND-USD",
						Price:     1e10,
						Timestamp: suite.ctx.BlockTime().Unix(),
						Status:    types.PriceStatusAvailable,
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
				Prices: []types.Price{
					{
						SignalID:  "CS:ATOM-USD",
						Price:     285171,
						Timestamp: suite.ctx.BlockTime().Unix(),
						Status:    types.PriceStatusAvailable,
					},
					{
						SignalID:  "CS:BAND-USD",
						Price:     285171,
						Timestamp: suite.ctx.BlockTime().Unix(),
						Status:    types.PriceStatusAvailable,
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
					Price:     0,
					Timestamp: 0,
					Status:    types.PriceStatusNotInCurrentFeeds,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     285171,
					Timestamp: suite.ctx.BlockTime().Unix(),
					Status:    types.PriceStatusAvailable,
				},
			},
			encoder: types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{
				Prices: []types.Price{
					{
						SignalID:  "CS:ATOM-USD",
						Price:     0,
						Timestamp: 0,
						Status:    types.PriceStatusNotInCurrentFeeds,
					},
					{
						SignalID:  "CS:BAND-USD",
						Price:     285171,
						Timestamp: suite.ctx.BlockTime().Unix(),
						Status:    types.PriceStatusAvailable,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
		},
		{
			name:      "fail case - price not found",
			signalIDs: []string{"CS:ATOM-USDfake"},
			setPrices: []types.Price{},
			encoder:   types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.FeedsPriceData{
				Prices: []types.Price{
					{
						SignalID:  "CS:ATOM-USDfake",
						Price:     0,
						Timestamp: 0,
						Status:    types.PriceStatusNotInCurrentFeeds,
					},
				},
				Timestamp: uint64(suite.ctx.BlockTime().Unix()),
			},
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
		})
	}
}
