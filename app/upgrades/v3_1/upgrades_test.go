package v3_1_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/upgrades"
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
	bandtesting.SetCustomUpgrades([]upgrades.Upgrade{v3_rc3.Upgrade})

	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)
}

// TestUpgrade ensures the test that does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	preUpgradeChecks(s)

	upgradeHeight := int64(2)
	s.ConfirmUpgradeSucceeded(v3_rc3.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
	// Set reward percentage of bandtss params to 2 to test upgrade
	// this is to ensure that the upgrade handler is called
	bandtssParams := s.app.BandtssKeeper.GetParams(s.ctx)
	bandtssParams.RewardPercentage = uint64(2)
	err := s.app.BandtssKeeper.SetParams(s.ctx, bandtssParams)
	s.Require().NoError(err)

	// check param is set to 2
	bandtssParams = s.app.BandtssKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(2), bandtssParams.RewardPercentage)
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// check bandtss params is changed after upgrade
	bandtssParams := s.app.BandtssKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(1), bandtssParams.RewardPercentage)
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
