package v3_mainnet_test

import (
	"testing"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	v3 "github.com/bandprotocol/chain/v3/app/upgrades/v3_mainnet"
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
	s.ConfirmUpgradeSucceeded(v3.UpgradeName, upgradeHeight)

	postUpgradeChecks(s)
}

func preUpgradeChecks(s *UpgradeTestSuite) {
}

func postUpgradeChecks(s *UpgradeTestSuite) {
	// check the subspaces
	for _, subspace := range s.app.ParamsKeeper.GetSubspaces() {
		s.Require().True(subspace.HasKeyTable())
	}

	// Check consensus params after upgrade
	consensusParam, err := s.app.ConsensusParamsKeeper.ParamsStore.Get(s.ctx)
	s.Require().NoError(err)

	s.Require().Equal(v3.BlockMaxBytes, consensusParam.Block.MaxBytes)
	s.Require().Equal(v3.BlockMaxGas, consensusParam.Block.MaxGas)
	s.Require().Equal([]string{cmttypes.ABCIPubKeyTypeSecp256k1}, consensusParam.Validator.PubKeyTypes)

	DefaultEvidenceParams := cmttypes.DefaultEvidenceParams()
	s.Require().Equal(DefaultEvidenceParams.MaxAgeNumBlocks, consensusParam.Evidence.MaxAgeNumBlocks)
	s.Require().Equal(DefaultEvidenceParams.MaxAgeDuration, consensusParam.Evidence.MaxAgeDuration)
	s.Require().Equal(DefaultEvidenceParams.MaxBytes, consensusParam.Evidence.MaxBytes)

	s.Require().Equal(cmttypes.DefaultVersionParams().App, consensusParam.Version.App)
	s.Require().
		Equal(cmttypes.DefaultABCIParams().VoteExtensionsEnableHeight, consensusParam.Abci.VoteExtensionsEnableHeight)

	// check ICA host params
	icaHostParams := s.app.ICAHostKeeper.GetParams(s.ctx)
	s.Require().True(icaHostParams.HostEnabled)
	s.Require().Equal(v3.ICAAllowMessages, icaHostParams.AllowMessages)

	// check feemarket params
	feemarketParams, err := s.app.FeeMarketKeeper.GetParams(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(feemarkettypes.DefaultWindow, feemarketParams.Window)
	s.Require().Equal(feemarkettypes.DefaultAlpha, feemarketParams.Alpha)
	s.Require().Equal(feemarkettypes.DefaultBeta, feemarketParams.Beta)
	s.Require().Equal(feemarkettypes.DefaultGamma, feemarketParams.Gamma)
	s.Require().Equal(feemarkettypes.DefaultMinLearningRate, feemarketParams.MinLearningRate)
	s.Require().Equal(feemarkettypes.DefaultMaxLearningRate, feemarketParams.MaxLearningRate)
	s.Require().Equal(uint64(v3.BlockMaxGas), feemarketParams.MaxBlockUtilization)
	s.Require().Equal(v3.MinimumGasPrice, feemarketParams.MinBaseGasPrice)
	s.Require().Equal(v3.Denom, feemarketParams.FeeDenom)
	s.Require().False(feemarketParams.Enabled)

	// check feemarket state
	state, err := s.app.FeeMarketKeeper.GetState(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(v3.MinimumGasPrice, state.BaseGasPrice)
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
