package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetParams() {
	ctx := suite.ctx

	expectedParams := types.DefaultParams()
	err := suite.feedsKeeper.SetParams(ctx, expectedParams)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedParams, suite.feedsKeeper.GetParams(ctx))
}
