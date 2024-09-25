package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

var defaultVotes = []abci.VoteInfo{{
	Validator: abci.Validator{
		Address: valConsPk0.Address(),
		Power:   70,
	},
}, {
	Validator: abci.Validator{
		Address: valConsPk1.Address(),
		Power:   20,
	},
}, {
	Validator: abci.Validator{
		Address: valConsPk2.Address(),
		Power:   10,
	},
}}

func (suite *KeeperTestSuite) mockValidators() {
	suite.stakingKeeper.EXPECT().
		ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).
		Return(validators[0].Validator, nil).AnyTimes()
	suite.stakingKeeper.EXPECT().
		ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).
		Return(validators[1].Validator, nil).AnyTimes()
	suite.stakingKeeper.EXPECT().
		ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).
		Return(validators[2].Validator, nil).AnyTimes()
}

func (suite *KeeperTestSuite) mockFundCommunityPool(amount sdk.Coins, sender sdk.AccAddress) {
	suite.distrKeeper.EXPECT().FundCommunityPool(
		gomock.Any(),
		amount, sender,
	)
}

func (suite *KeeperTestSuite) TestAllocateTokenNoActiveValidators() {
	ctx := suite.ctx
	k := suite.oracleKeeper

	// No active oracle validators so nothing should happen.
	// Expect only try to sum of validator power
	suite.mockValidators()
	k.AllocateTokens(ctx, defaultVotes)
}

func (suite *KeeperTestSuite) TestAllocateTokensOneActive() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	suite.mockValidators()
	// Set Oracle param for reward percentage
	params := types.DefaultParams()
	params.OracleRewardPercentage = 70
	err := k.SetParams(ctx, params)
	require.NoError(err)

	err = k.Activate(ctx, validators[1].Address)
	require.NoError(err)

	// From 70% of fee, 2% should go to community pool, the rest goes to the only active validator.
	// Mock all keeper that will be called when allocate token
	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc)
	suite.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(coins1000000uband)

	suite.bankKeeper.EXPECT().
		SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", distrtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 700000)))

	suite.distrKeeper.EXPECT().GetCommunityTax(gomock.Any()).Return(math.LegacyNewDecWithPrec(2, 2), nil)
	suite.distrKeeper.EXPECT().
		AllocateTokensToValidator(gomock.Any(), validators[1].Validator, sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(686000)}})
	suite.distrKeeper.EXPECT().
		AllocateTokensToValidator(gomock.Any(), validators[1].Validator, (sdk.DecCoins)(nil))
	distAcc := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)
	suite.mockFundCommunityPool(sdk.NewCoins(sdk.NewInt64Coin("uband", 14000)), distAcc.GetAddress())
	suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), distrtypes.ModuleName).Return(distAcc)
	suite.stakingKeeper.EXPECT().Validator(gomock.Any(), validators[1].Address).Return(validators[1].Validator, nil)

	// Set validator 1 as the proposer
	ctx = ctx.WithProposer(validators[1].Address.Bytes())

	k.AllocateTokens(ctx, defaultVotes)
}

// func (suite *KeeperTestSuite) TestAllocateTokensAllActive() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	feeCollector := SetupFeeCollector(app.BandApp, ctx, k)

// 	require.Equal(t, Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
// 	// From 70% of fee, 2% should go to community pool, the rest get split to validators.
// 	k.AllocateTokens(ctx, defaultVotes())

// 	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
// 	require.Equal(
// 		t,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 300000)),
// 		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
// 	)
// 	require.Equal(
// 		t,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 700000)),
// 		app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()),
// 	)

// 	// Check fee pool
// 	feePool, err := app.DistrKeeper.FeePool.Get(ctx)
// 	require.NoError(t, err)

// 	require.Equal(
// 		t,
// 		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(14000)}},
// 		feePool.CommunityPool,
// 	)

// 	rewards0, err := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress)
// 	require.NoError(t, err)
// 	require.Equal(
// 		t,
// 		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(480200)}},
// 		rewards0.Rewards,
// 	)

// 	rewards1, err := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
// 	require.NoError(t, err)
// 	require.Equal(
// 		t,
// 		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(137200)}},
// 		rewards1.Rewards,
// 	)

// 	rewards2, err := app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[2].ValAddress)
// 	require.NoError(t, err)
// 	require.Equal(
// 		t,
// 		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(68600)}},
// 		rewards2.Rewards,
// 	)
// }

// func (suite *KeeperTestSuite) TestGetDefaultValidatorStatus() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
// }

// func (suite *KeeperTestSuite) TestGetSetValidatorStatus() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	now := time.Now().UTC()
// 	// After setting status of the 1st validator, we should be able to get it back.
// 	k.SetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(true, now), vs)
// 	vs = k.GetValidatorStatus(ctx, bandtesting.Validators[1].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
// }

// func (suite *KeeperTestSuite) TestActivateValidatorOK() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	now := time.Now().UTC()
// 	ctx = ctx.WithBlockTime(now)
// 	err := k.Activate(ctx, bandtesting.Validators[0].ValAddress)
// 	require.NoError(t, err)
// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(true, now), vs)
// 	vs = k.GetValidatorStatus(ctx, bandtesting.Validators[1].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
// }

// func (suite *KeeperTestSuite) TestFailActivateAlreadyActive() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	now := time.Now().UTC()
// 	ctx = ctx.WithBlockTime(now)
// 	err := k.Activate(ctx, bandtesting.Validators[0].ValAddress)
// 	require.NoError(t, err)
// 	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
// 	require.ErrorIs(t, err, types.ErrValidatorAlreadyActive)
// }

// func (suite *KeeperTestSuite) TestFailActivateTooSoon() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	now := time.Now().UTC()
// 	// Set validator to be inactive just now.
// 	k.SetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress, types.NewValidatorStatus(false, now))
// 	// You can't activate until it's been at least InactivePenaltyDuration nanosec.
// 	penaltyDuration := k.GetParams(ctx).InactivePenaltyDuration
// 	require.ErrorIs(
// 		t,
// 		k.Activate(ctx.WithBlockTime(now), bandtesting.Validators[0].ValAddress),
// 		types.ErrTooSoonToActivate,
// 	)
// 	require.ErrorIs(
// 		t,
// 		k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration/2))), bandtesting.Validators[0].ValAddress),
// 		types.ErrTooSoonToActivate,
// 	)
// 	// So far there must be no changes to the validator's status.
// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(false, now), vs)
// 	// Now the time has come.
// 	require.NoError(
// 		t,
// 		k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration))), bandtesting.Validators[0].ValAddress),
// 	)
// 	vs = k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(true, now.Add(time.Duration(penaltyDuration))), vs)
// }

// func (suite *KeeperTestSuite) TestMissReportSuccess() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	now := time.Now().UTC()
// 	next := now.Add(time.Duration(10))
// 	k.SetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
// 	k.MissReport(ctx.WithBlockTime(next), bandtesting.Validators[0].ValAddress, next)
// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(false, next), vs)
// }

// func (suite *KeeperTestSuite) TestMissReportTooSoonNoop() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	prev := time.Now().UTC()
// 	now := prev.Add(time.Duration(10))
// 	k.SetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
// 	k.MissReport(ctx.WithBlockTime(prev), bandtesting.Validators[0].ValAddress, prev)
// 	vs := k.GetValidatorStatus(ctx, bandtesting.Validators[0].ValAddress)
// 	require.Equal(t, types.NewValidatorStatus(true, now), vs)
// }

func (suite *KeeperTestSuite) TestMissReportAlreadyInactiveNoop() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	next := now.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, validators[0].Address, types.NewValidatorStatus(false, now))
	k.MissReport(ctx.WithBlockTime(next), validators[0].Address, next)
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(false, now), vs)
}
