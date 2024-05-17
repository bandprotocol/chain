package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetDeletePrice() {
	ctx := suite.ctx

	// set
	expPrice := types.Price{
		SignalID:  "crypto_price.bandusd",
		Price:     1e10,
		Timestamp: ctx.BlockTime().Unix(),
	}
	suite.feedsKeeper.SetPrice(ctx, expPrice)

	// get
	price, err := suite.feedsKeeper.GetPrice(ctx, "crypto_price.bandusd")
	suite.Require().NoError(err)
	suite.Require().Equal(expPrice, price)

	// delete
	suite.feedsKeeper.DeletePrice(ctx, "crypto_price.bandusd")

	// get
	_, err = suite.feedsKeeper.GetPrice(ctx, "crypto_price.bandusd")
	suite.Require().ErrorContains(err, "price not found")
}

func (suite *KeeperTestSuite) TestGetSetPrices() {
	ctx := suite.ctx

	// set
	expPrices := []types.Price{
		{
			SignalID:  "crypto_price.atomusd",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			SignalID:  "crypto_price.bandusd",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	suite.feedsKeeper.SetPrices(ctx, expPrices)

	// get
	prices := suite.feedsKeeper.GetPrices(ctx)
	suite.Require().Equal(expPrices, prices)
}

func (suite *KeeperTestSuite) TestGetSetDeleteValidatorPrice() {
	ctx := suite.ctx

	// set
	expPriceVal := types.ValidatorPrice{
		Validator: ValidValidator.String(),
		SignalID:  "crypto_price.bandusd",
		Price:     1e10,
		Timestamp: ctx.BlockTime().Unix(),
	}
	err := suite.feedsKeeper.SetValidatorPrice(ctx, expPriceVal)
	suite.Require().NoError(err)

	// get
	priceVal, err := suite.feedsKeeper.GetValidatorPrice(ctx, "crypto_price.bandusd", ValidValidator)
	suite.Require().NoError(err)
	suite.Require().Equal(expPriceVal, priceVal)

	// delete
	suite.feedsKeeper.DeleteValidatorPrice(ctx, "crypto_price.bandusd", ValidValidator)

	// get
	_, err = suite.feedsKeeper.GetValidatorPrice(ctx, "crypto_price.bandusd", ValidValidator)
	suite.Require().ErrorContains(err, "validator price not found")
}

func (suite *KeeperTestSuite) TestGetSetValidatorPrices() {
	ctx := suite.ctx

	// set
	expPriceVals := []types.ValidatorPrice{
		{
			Validator: ValidValidator.String(),
			SignalID:  "crypto_price.bandusd",
			Price:     1e10,
			Timestamp: ctx.BlockTime().Unix(),
		},
		{
			Validator: ValidValidator2.String(),
			SignalID:  "crypto_price.bandusd",
			Price:     1e10 + 5,
			Timestamp: ctx.BlockTime().Unix(),
		},
	}
	err := suite.feedsKeeper.SetValidatorPrices(ctx, expPriceVals)
	suite.Require().NoError(err)

	// get
	priceVals := suite.feedsKeeper.GetValidatorPrices(ctx, "crypto_price.bandusd")
	suite.Require().Equal(expPriceVals, priceVals)
}

func (suite *KeeperTestSuite) TestCalculatePrice() {
	ctx := suite.ctx

	// set
	feed := types.Feed{
		SignalID:              "crypto_price.bandusd",
		Interval:              60,
		DeviationInThousandth: 5,
	}
	suite.feedsKeeper.SetSupportedFeeds(ctx, []types.Feed{feed})

	err := suite.feedsKeeper.SetValidatorPrices(ctx, []types.ValidatorPrice{
		{
			PriceStatus: types.PriceStatusAvailable,
			Validator:   ValidValidator.String(),
			SignalID:    "crypto_price.bandusd",
			Price:       1000,
			Timestamp:   ctx.BlockTime().Unix(),
		},
		{
			PriceStatus: types.PriceStatusAvailable,
			Validator:   ValidValidator2.String(),
			SignalID:    "crypto_price.bandusd",
			Price:       2000,
			Timestamp:   ctx.BlockTime().Unix(),
		},
	})
	suite.Require().NoError(err)

	// cal
	price, err := suite.feedsKeeper.CalculatePrice(ctx, feed, ctx.BlockTime().Unix(), ctx.BlockHeight())
	suite.Require().NoError(err)
	suite.Require().Equal(types.Price{
		PriceStatus: types.PriceStatusAvailable,
		SignalID:    "crypto_price.bandusd",
		Price:       1000,
		Timestamp:   ctx.BlockTime().Unix(),
	}, price)
}
