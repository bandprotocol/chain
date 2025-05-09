package v3_rc4_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/upgrades/v3_rc4"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
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
	s.ConfirmUpgradeSucceeded(v3_rc4.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// Verify changes made by the upgrade
	acc, perms := s.app.AccountKeeper.GetModuleAccountAndPermissions(s.ctx, tunneltypes.ModuleName)
	s.Require().NotNil(acc)
	s.Require().Contains(acc.GetPermissions(), authtypes.Minter)
	s.Require().Equal(perms, acc.GetPermissions())
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
