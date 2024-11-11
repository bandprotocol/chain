package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetReferenceSourceConfig() {
	ctx := suite.ctx

	// set
	expReferenceSourceConfig := types.ReferenceSourceConfig{
		RegistryIPFSHash: "hash",
		RegistryVersion:  "1.0.0",
	}
	err := suite.feedsKeeper.SetReferenceSourceConfig(ctx, expReferenceSourceConfig)
	suite.Require().NoError(err)

	// get
	referenceSourceConfig := suite.feedsKeeper.GetReferenceSourceConfig(ctx)
	suite.Require().Equal(expReferenceSourceConfig, referenceSourceConfig)
}
