package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetDeletePrice() {
	ctx := suite.ctx

	// set
	expPrice := types.Price{
		SignalID:  "CS:BAND-USD",
		Price:     1e10,
		Timestamp: ctx.BlockTime().Unix(),
	}
	suite.feedsKeeper.SetPrice(ctx, expPrice)

	// get
	price, err := suite.feedsKeeper.GetPrice(ctx, "CS:BAND-USD")
	suite.Require().NoError(err)
	suite.Require().Equal(expPrice, price)
}

func (suite *KeeperTestSuite) TestGetSetPrices() {
	ctx := suite.ctx

	// set
	expPrices := []types.Price{
		{
			SignalID:  "CS:ATOM-USD",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			SignalID:  "CS:BAND-USD",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	suite.feedsKeeper.SetPrices(ctx, expPrices)

	// get
	prices := suite.feedsKeeper.GetPrices(ctx)
	suite.Require().Equal(expPrices, prices)
}

func (suite *KeeperTestSuite) TestGetSetValidatorPriceList() {
	ctx := suite.ctx

	// set
	expValPrices := []types.ValidatorPrice{
		{
			Validator: ValidValidator.String(),
			SignalID:  "CS:BAND-USD",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			Validator: ValidValidator.String(),
			SignalID:  "CS:ETH-USD",
			Price:     1e10 + 5,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	err := suite.feedsKeeper.SetValidatorPriceList(ctx, ValidValidator, expValPrices)
	suite.Require().NoError(err)

	// get
	valPrices, err := suite.feedsKeeper.GetValidatorPriceList(ctx, ValidValidator)
	suite.Require().NoError(err)
	suite.Require().Equal(expValPrices, valPrices.ValidatorPrices)
}

func (suite *KeeperTestSuite) TestCalculatePrice() {
	ctx := suite.ctx

	// set
	feed := types.Feed{
		SignalID: "CS:BAND-USD",
		Interval: 60,
	}
	priceFeedInfos := []types.PriceFeedInfo{
		{
			PriceStatus: types.PriceStatusAvailable,
			Power:       5000,
			Price:       1000,
			Timestamp:   1719914474,
			Index:       0,
		},
		{
			PriceStatus: types.PriceStatusAvailable,
			Power:       3000,
			Price:       2000,
			Timestamp:   1719914474,
			Index:       1,
		},
		{
			PriceStatus: types.PriceStatusAvailable,
			Power:       3000,
			Price:       2000,
			Timestamp:   1719914474,
			Index:       2,
		},
	}
	price, err := suite.feedsKeeper.CalculatePrice(
		ctx,
		feed,
		priceFeedInfos,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(types.Price{
		PriceStatus: types.PriceStatusAvailable,
		SignalID:    "CS:BAND-USD",
		Price:       1000,
		Timestamp:   ctx.BlockTime().Unix(),
	}, price)
}
