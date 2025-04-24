package v3_rc3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/upgrades/v3_rc3"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
)

type UpgradeTestSuite struct {
	suite.Suite

	app *band.BandApp
	ctx sdk.Context
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupTest() {
	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

	// Activate validators
	for _, v := range bandtesting.Validators {
		err := s.app.OracleKeeper.Activate(s.ctx, v.ValAddress)
		s.Require().NoError(err)
	}

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	preUpgradeChecks(s)

	upgradeHeight := int64(2)
	s.ConfirmUpgradeSucceeded(v3_rc3.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	// check default oracle params
	oracleParams := s.app.OracleKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(512), oracleParams.MaxCalldataSize)
	s.Require().Equal(uint64(512), oracleParams.MaxReportDataSize)

	// Set oracle params to 1 to test upgrade
	// this is to ensure that the upgrade handler is called
	oracleParams.MaxCalldataSize = uint64(1)
	oracleParams.MaxReportDataSize = uint64(1)
	err := s.app.OracleKeeper.SetParams(s.ctx, oracleParams)
	s.Require().NoError(err)

	// check oracle params is set to 1
	oracleParams = s.app.OracleKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(1), oracleParams.MaxCalldataSize)
	s.Require().Equal(uint64(1), oracleParams.MaxReportDataSize)
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// check oracle params is changed after upgrade
	oracleParams := s.app.OracleKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(512), oracleParams.MaxCalldataSize)
	s.Require().Equal(uint64(512), oracleParams.MaxReportDataSize)
}

func (s *UpgradeTestSuite) ConfirmUpgradeSucceeded(upgradeName string, upgradeHeight int64) {
	plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
	err := s.app.AppKeepers.UpgradeKeeper.ScheduleUpgrade(s.ctx, plan)
	s.Require().NoError(err)
	_, err = s.app.AppKeepers.UpgradeKeeper.GetUpgradePlan(s.ctx)
	s.Require().NoError(err)

	s.ctx = s.ctx.WithBlockHeight(upgradeHeight)
	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.ctx.BlockHeight()})
	s.Require().NoError(err)
}
