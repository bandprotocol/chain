package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestGetSetParams() {
	ctx := suite.ctx

	expectedParams := types.DefaultParams()

	// set
	err := suite.restakeKeeper.SetParams(ctx, expectedParams)
	suite.Require().NoError(err)

	// get
	suite.Require().Equal(expectedParams, suite.restakeKeeper.GetParams(ctx))

	// set invalid params
	err = suite.restakeKeeper.SetParams(ctx, types.Params{
		AllowedDenoms: []string{""},
	})
	suite.Require().Error(err)

	// get after set invalid params - params should not be changed.
	suite.Require().Equal(expectedParams, suite.restakeKeeper.GetParams(ctx))
}
