package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/restake/types"
)

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.ctx
	suite.setupState()

	exportGenesis := suite.restakeKeeper.ExportGenesis(ctx)

	suite.Require().Equal(suite.validVaults, exportGenesis.Vaults)
	suite.Require().Equal(suite.validLocks, exportGenesis.Locks)
	suite.Require().Equal(suite.validStakes, exportGenesis.Stakes)
	suite.Require().Equal(suite.validParams, exportGenesis.Params)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	ctx := suite.ctx

	suite.bankKeeper.EXPECT().
		GetAllBalances(gomock.Any(), suite.restakeKeeper.GetModuleAccount(ctx).GetAddress()).
		Return(sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(60)))).
		Times(1)

	g := types.NewGenesisState(suite.validParams, suite.validVaults, suite.validLocks, suite.validStakes)
	suite.restakeKeeper.InitGenesis(suite.ctx, g)

	suite.Require().Equal(suite.validVaults, suite.restakeKeeper.GetVaults(ctx))
	suite.Require().Equal(suite.validLocks, suite.restakeKeeper.GetLocks(ctx))
	suite.Require().Equal(suite.validStakes, suite.restakeKeeper.GetStakes(ctx))
	suite.Require().Equal(suite.validParams, suite.restakeKeeper.GetParams(ctx))
}
