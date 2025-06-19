package v3_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/upgrades"
	v3 "github.com/bandprotocol/chain/v3/app/upgrades/v3"
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
	bandtesting.SetCustomUpgrades([]upgrades.Upgrade{v3.Upgrade})

	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

	// Activate validators
	for _, v := range bandtesting.Validators {
		err := s.app.OracleKeeper.Activate(s.ctx, v.ValAddress)
		s.Require().NoError(err)

		reporter := sdk.AccAddress("1000000001")
		expUnix := time.Unix(32518321013, 0)

		err = s.app.AuthzKeeper.SaveGrant(s.ctx, v.Address, reporter, authz.NewGenericAuthorization("/oracle.v1.MsgReportData"), &expUnix)
		s.Require().NoError(err)
	}

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)
}

// TestUpgrade ensures the test that does not error out.
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

	// check oracle params
	oracleParams := s.app.OracleKeeper.GetParams(s.ctx)
	s.Require().Equal(uint64(512), oracleParams.MaxCalldataSize)
	s.Require().Equal(uint64(512), oracleParams.MaxReportDataSize)

	// check global fee params
	s.Require().
		Equal(sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(25, 4))}, s.app.GlobalFeeKeeper.GetParams(s.ctx).MinimumGasPrices)

	// check gov params
	govParams, err := s.app.GovKeeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().
		Equal(sdk.Coins{sdk.NewCoin("uband", sdkmath.NewInt(5000*1000000))}, sdk.Coins(govParams.ExpeditedMinDeposit))
	s.Require().
		Equal(1*24*time.Hour, *govParams.ExpeditedVotingPeriod)
	s.Require().
		Equal(5*24*time.Hour, *govParams.MaxDepositPeriod)
	s.Require().
		Equal(5*24*time.Hour, *govParams.VotingPeriod)

	for _, v := range bandtesting.Validators {
		reporter := sdk.AccAddress("1000000001")

		grants, err := s.app.AuthzKeeper.Grants(s.ctx, &authz.QueryGrantsRequest{
			Granter: v.Address.String(),
			Grantee: reporter.String(),
		})
		s.Require().NoError(err)

		for _, grant := range grants.Grants {
			auth, err := grant.GetAuthorization()
			s.Require().NoError(err)
			s.Require().Equal("/band.oracle.v1.MsgActivate", auth.MsgTypeURL())
		}
	}
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
