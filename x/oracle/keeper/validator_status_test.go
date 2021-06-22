package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func defaultVotes() []abci.VoteInfo {
	return []abci.VoteInfo{{
		Validator: abci.Validator{
			Address: testapp.Validators[0].PubKey.Address(),
			Power:   70,
		},
		SignedLastBlock: true,
	}, {
		Validator: abci.Validator{
			Address: testapp.Validators[1].PubKey.Address(),
			Power:   20,
		},
		SignedLastBlock: true,
	}, {
		Validator: abci.Validator{
			Address: testapp.Validators[2].PubKey.Address(),
			Power:   10,
		},
		SignedLastBlock: true,
	}}
}

func SetupFeeCollector(app *testapp.TestingApp, ctx sdk.Context, k keeper.Keeper) authtypes.ModuleAccountI {
	// Set collected fee to 1000000uband and 70% oracle reward proportion.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	app.BankKeeper.AddCoins(ctx, feeCollector.GetAddress(), Coins1000000uband)
	app.AccountKeeper.SetAccount(ctx, feeCollector)

	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 70
	k.SetParams(ctx, params)

	return feeCollector
}

func TestAllocateTokenNoActiveValidators(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(false)
	feeCollector := SetupFeeCollector(app, ctx, k)

	require.Equal(t, Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// No active oracle validators so nothing should happen.
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	require.Equal(t, Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	require.Empty(t, app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
}

func TestAllocateTokensOneActive(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(false)
	feeCollector := SetupFeeCollector(app, ctx, k)

	require.Equal(t, Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// From 70% of fee, 2% should go to community pool, the rest goes to the only active validator.
	k.Activate(ctx, testapp.Validators[1].ValAddress)
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uband", 300000)), app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uband", 700000)), app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(14000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	require.Empty(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(686000)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress).Rewards)
	require.Empty(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[2].ValAddress))
}

func TestAllocateTokensAllActive(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)
	feeCollector := SetupFeeCollector(app, ctx, k)

	require.Equal(t, Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// From 70% of fee, 2% should go to community pool, the rest get split to validators.
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uband", 300000)), app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("uband", 700000)), app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(14000)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(480200)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(137200)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress).Rewards)
	require.Equal(t, sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(68600)}}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[2].ValAddress).Rewards)
}

func TestGetDefaultValidatorStatus(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
}

func TestGetSetValidatorStatus(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	// After setting status of the 1st validator, we should be able to get it back.
	k.SetValidatorStatus(ctx, testapp.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(true, now), vs)
	vs = k.GetValidatorStatus(ctx, testapp.Validators[1].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
}

func TestActivateValidatorOK(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	err := k.Activate(ctx, testapp.Validators[0].ValAddress)
	require.NoError(t, err)
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(true, now), vs)
	vs = k.GetValidatorStatus(ctx, testapp.Validators[1].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, time.Time{}), vs)
}

func TestFailActivateAlreadyActive(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	ctx = ctx.WithBlockTime(now)
	err := k.Activate(ctx, testapp.Validators[0].ValAddress)
	require.NoError(t, err)
	err = k.Activate(ctx, testapp.Validators[0].ValAddress)
	require.ErrorIs(t, err, types.ErrValidatorAlreadyActive)
}

func TestFailActivateTooSoon(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	// Set validator to be inactive just now.
	k.SetValidatorStatus(ctx, testapp.Validators[0].ValAddress, types.NewValidatorStatus(false, now))
	// You can't activate until it's been at least InactivePenaltyDuration nanosec.
	penaltyDuration := k.GetParams(ctx).InactivePenaltyDuration
	require.ErrorIs(t, k.Activate(ctx.WithBlockTime(now), testapp.Validators[0].ValAddress), types.ErrTooSoonToActivate)
	require.ErrorIs(t, k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration/2))), testapp.Validators[0].ValAddress), types.ErrTooSoonToActivate)
	// So far there must be no changes to the validator's status.
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, now), vs)
	// Now the time has come.
	require.NoError(t, k.Activate(ctx.WithBlockTime(now.Add(time.Duration(penaltyDuration))), testapp.Validators[0].ValAddress))
	vs = k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(true, now.Add(time.Duration(penaltyDuration))), vs)
}

func TestMissReportSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	next := now.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, testapp.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
	k.MissReport(ctx.WithBlockTime(next), testapp.Validators[0].ValAddress, next)
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, next), vs)
}

func TestMissReportTooSoonNoop(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	prev := time.Now().UTC()
	now := prev.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, testapp.Validators[0].ValAddress, types.NewValidatorStatus(true, now))
	k.MissReport(ctx.WithBlockTime(prev), testapp.Validators[0].ValAddress, prev)
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(true, now), vs)
}

func TestMissReportAlreadyInactiveNoop(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	now := time.Now().UTC()
	next := now.Add(time.Duration(10))
	k.SetValidatorStatus(ctx, testapp.Validators[0].ValAddress, types.NewValidatorStatus(false, now))
	k.MissReport(ctx.WithBlockTime(next), testapp.Validators[0].ValAddress, next)
	vs := k.GetValidatorStatus(ctx, testapp.Validators[0].ValAddress)
	require.Equal(t, types.NewValidatorStatus(false, now), vs)
}
