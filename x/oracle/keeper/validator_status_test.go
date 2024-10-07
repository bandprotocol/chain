package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	oracletestutil "github.com/bandprotocol/chain/v3/x/oracle/testutil"
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

func (suite *KeeperTestSuite) mockValidators() []*gomock.Call {
	return []*gomock.Call{
		suite.stakingKeeper.EXPECT().
			ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk0)).
			Return(validators[0].Validator, nil).AnyTimes(),
		suite.stakingKeeper.EXPECT().
			ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk1)).
			Return(validators[1].Validator, nil).AnyTimes(),
		suite.stakingKeeper.EXPECT().
			ValidatorByConsAddr(gomock.Any(), sdk.GetConsAddress(valConsPk2)).
			Return(validators[2].Validator, nil).AnyTimes(),
	}
}

func (suite *KeeperTestSuite) mockFundCommunityPool(amount sdk.Coins, sender sdk.AccAddress) *gomock.Call {
	return suite.distrKeeper.EXPECT().FundCommunityPool(
		gomock.Any(),
		amount, sender,
	)
}

func (suite *KeeperTestSuite) TestAllocateTokenNoActiveValidators() {
	ctx := suite.ctx
	k := suite.oracleKeeper

	// No active oracle validators so nothing should happen.
	// Expect only try to sum of validator power
	oracletestutil.ChainGoMockCalls(suite.mockValidators()...)
	k.AllocateTokens(ctx, defaultVotes)
}

func (suite *KeeperTestSuite) TestAllocateTokensOneActive() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Set Oracle param for reward percentage
	params := types.DefaultParams()
	params.OracleRewardPercentage = 70
	err := k.SetParams(ctx, params)
	require.NoError(err)

	err = k.Activate(ctx, validators[1].Address)
	require.NoError(err)

	// Set validator 1 as the proposer
	ctx = ctx.WithProposer(valConsPk1.Address().Bytes())

	// Mock all keeper that will be called when allocate token by order
	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	distAcc := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)

	oracletestutil.ChainGoMockCalls(
		oracletestutil.ChainGoMockCalls(suite.mockValidators()...),
		suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc),
		suite.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(coins1000000uband),
		suite.bankKeeper.EXPECT().
			SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", distrtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 700000))),
		suite.distrKeeper.EXPECT().GetCommunityTax(gomock.Any()).Return(math.LegacyNewDecWithPrec(2, 2), nil),
		suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), distrtypes.ModuleName).Return(distAcc),
		suite.mockFundCommunityPool(sdk.NewCoins(sdk.NewInt64Coin("uband", 14000)), distAcc.GetAddress()),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[1].Validator, sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(686000)}}),
		suite.stakingKeeper.EXPECT().
			ValidatorByConsAddr(gomock.Any(), valConsPk1.Address().Bytes()).
			Return(validators[1].Validator, nil),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[1].Validator, (sdk.DecCoins)(nil)),
	)

	k.AllocateTokens(ctx, defaultVotes)
}

func (suite *KeeperTestSuite) TestAllocateTokensAllActive() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Set Oracle param for reward percentage
	params := types.DefaultParams()
	params.OracleRewardPercentage = 70
	err := k.SetParams(ctx, params)
	require.NoError(err)

	err = k.Activate(ctx, validators[0].Address)
	require.NoError(err)
	err = k.Activate(ctx, validators[1].Address)
	require.NoError(err)
	err = k.Activate(ctx, validators[2].Address)
	require.NoError(err)

	// Set validator 0 as the proposer
	ctx = ctx.WithProposer(valConsPk0.Address().Bytes())

	// Mock all keeper that will be called when allocate token by order
	feeCollectorAcc := authtypes.NewEmptyModuleAccount("fee_collector")
	distAcc := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)
	oracletestutil.ChainGoMockCalls(
		oracletestutil.ChainGoMockCalls(suite.mockValidators()...),
		suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), "fee_collector").Return(feeCollectorAcc),
		suite.bankKeeper.EXPECT().GetAllBalances(gomock.Any(), feeCollectorAcc.GetAddress()).Return(coins1000000uband),
		suite.bankKeeper.EXPECT().
			SendCoinsFromModuleToModule(gomock.Any(), "fee_collector", distrtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 700000))),
		suite.distrKeeper.EXPECT().GetCommunityTax(gomock.Any()).Return(math.LegacyNewDecWithPrec(2, 2), nil),
		suite.authKeeper.EXPECT().GetModuleAccount(gomock.Any(), distrtypes.ModuleName).Return(distAcc),
		suite.mockFundCommunityPool(sdk.NewCoins(sdk.NewInt64Coin("uband", 14000)), distAcc.GetAddress()),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[0].Validator, sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(480200)}}),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[1].Validator, sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(137200)}}),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[2].Validator, sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(68600)}}),
		suite.stakingKeeper.EXPECT().
			ValidatorByConsAddr(gomock.Any(), valConsPk0.Address().Bytes()).
			Return(validators[0].Validator, nil),
		suite.distrKeeper.EXPECT().
			AllocateTokensToValidator(gomock.Any(), validators[0].Validator, (sdk.DecCoins)(nil)),
	)

	k.AllocateTokens(ctx, defaultVotes)
}

func (suite *KeeperTestSuite) TestGetDefaultValidatorStatus() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(false, time.Time{}), vs)
}

func (suite *KeeperTestSuite) TestGetSetValidatorStatus() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	// After setting status of the 1st validator, we should be able to get it back.
	k.SetValidatorStatus(ctx, validators[0].Address, types.NewValidatorStatus(true, now))
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(true, now), vs)
	vs = k.GetValidatorStatus(ctx, validators[1].Address)
	require.Equal(types.NewValidatorStatus(false, time.Time{}), vs)
}

func (suite *KeeperTestSuite) TestActivateValidatorOK() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	err := k.Activate(ctx, validators[0].Address)
	require.NoError(err)
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(true, now), vs)
	vs = k.GetValidatorStatus(ctx, validators[1].Address)
	require.Equal(types.NewValidatorStatus(false, time.Time{}), vs)
}

func (suite *KeeperTestSuite) TestFailActivateAlreadyActive() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	err := k.Activate(ctx, validators[0].Address)
	require.NoError(err)
	err = k.Activate(ctx, validators[0].Address)
	require.ErrorIs(err, types.ErrValidatorAlreadyActive)
}

func (suite *KeeperTestSuite) TestFailActivateTooSoon() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	// Set validator to be inactive just now.
	k.SetValidatorStatus(ctx, validators[0].Address, types.NewValidatorStatus(false, now))
	// You can't activate until it's been at least InactivePenaltyDuration nanosec.
	penaltyDuration := k.GetParams(ctx).InactivePenaltyDuration
	require.ErrorIs(
		k.Activate(ctx.WithBlockTime(now), validators[0].Address),
		types.ErrTooSoonToActivate,
	)
	require.ErrorIs(
		k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration/2))), validators[0].Address),
		types.ErrTooSoonToActivate,
	)
	// So far there must be no changes to the validator's status.
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(false, now), vs)
	// Now the time has come.
	require.NoError(
		k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration))), validators[0].Address),
	)
	vs = k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(true, now.Add(time.Duration(penaltyDuration))), vs)
}

func (suite *KeeperTestSuite) TestMissReportSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	now := time.Now().UTC()
	next := now.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, validators[0].Address, types.NewValidatorStatus(true, now))
	k.MissReport(ctx.WithBlockTime(next), validators[0].Address, next)
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(false, next), vs)
}

func (suite *KeeperTestSuite) TestMissReportTooSoonNoop() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	prev := time.Now().UTC()
	now := prev.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, validators[0].Address, types.NewValidatorStatus(true, now))
	k.MissReport(ctx.WithBlockTime(prev), validators[0].Address, prev)
	vs := k.GetValidatorStatus(ctx, validators[0].Address)
	require.Equal(types.NewValidatorStatus(true, now), vs)
}

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
