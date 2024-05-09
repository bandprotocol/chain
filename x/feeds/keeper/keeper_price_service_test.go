package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetPriceService() {
	ctx := suite.ctx

	// set
	expPriceService := types.PriceService{
		Hash:    "hash",
		Version: "1.0.0",
		Url:     "https://bandprotocol.com/",
	}
	err := suite.feedsKeeper.SetPriceService(ctx, expPriceService)
	suite.Require().NoError(err)

	// get
	priceService := suite.feedsKeeper.GetPriceService(ctx)
	suite.Require().Equal(expPriceService, priceService)
}
