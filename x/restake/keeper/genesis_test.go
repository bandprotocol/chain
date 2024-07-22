package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/restake/types"
)

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.ctx
	suite.setupState()

	exportGenesis := suite.restakeKeeper.ExportGenesis(ctx)

	suite.Require().Equal(suite.validKeys, exportGenesis.Keys)
	suite.Require().Equal(suite.validLocks, exportGenesis.Locks)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	g := types.NewGenesisState(suite.validKeys, suite.validLocks)
	suite.restakeKeeper.InitGenesis(suite.ctx, g)

	suite.Require().Equal(suite.validKeys, suite.restakeKeeper.GetKeys(ctx))
	suite.Require().Equal(suite.validLocks, suite.restakeKeeper.GetLocks(ctx))
}
