package oracle_test

import (
	"encoding/hex"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
)

func fromHex(hexStr string) []byte {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return res
}

func TestRollingSeedCorrect(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	// Initially rolling seed should be all zeros.
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000000"),
		k.GetRollingSeed(ctx),
	)
	// Every begin block, the rolling seed should get updated.
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0100000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000001"),
		k.GetRollingSeed(ctx),
	)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0200000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(
		t,
		fromHex("0000000000000000000000000000000000000000000000000000000000000102"),
		k.GetRollingSeed(ctx),
	)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	})
	require.Equal(
		t,
		fromHex("00000000000000000000000000000000000000000000000000000000000102ff"),
		k.GetRollingSeed(ctx),
	)
}

func TestAllocateTokensCalledOnBeginBlock(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	votes := []abci.VoteInfo{{
		Validator:       abci.Validator{Address: bandtesting.Validators[0].PubKey.Address(), Power: 70},
		SignedLastBlock: true,
	}, {
		Validator:       abci.Validator{Address: bandtesting.Validators[1].PubKey.Address(), Power: 30},
		SignedLastBlock: true,
	}}
	// Set collected fee to 100uband + 70% oracle reward proportion + disable minting inflation.
	// NOTE: we intentionally keep ctx.BlockHeight = 0, so distr's AllocateTokens doesn't get called.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 100)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
	)
	require.NoError(t, err)

	distModule := app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)

	app.AccountKeeper.SetAccount(ctx, feeCollector)
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	err = app.MintKeeper.SetParams(ctx, mintParams)
	require.NoError(t, err)

	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 70
	err = k.SetParams(ctx, params)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	// If there are no validators active, Calling begin block should be no-op.
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.CommitInfo{Votes: votes},
	})
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	// 1 validator active, begin block should take 70% of the fee. 2% of that goes to comm pool.
	err = k.Activate(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(t, err)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.CommitInfo{Votes: votes},
	})
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 30)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 70)),
		app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 100*70%*2% = 1.4uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(14, 1)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	// 0uband
	require.Empty(t, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress))
	// 100*70%*98% = 68.6uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(686, 1)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress).Rewards,
	)
	// 2 validators active now. 70% of the remaining fee pool will be split 3 ways (comm pool + val1 + val2).
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(t, err)

	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.CommitInfo{Votes: votes},
	})
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 9)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 91)),
		app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 1.4uband + 30*70%*2% = 1.82uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(182, 2)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	// 30*70%*98%*70% = 14.406uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(14406, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress).Rewards,
	)
	// 68.6uband + 30*70%*98%*30% = 74.774uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(74774, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress).Rewards,
	)
	// 1 validator becomes in active, and will not get reward this time.
	k.MissReport(ctx, bandtesting.Validators[1].ValAddress, bandtesting.ParseTime(100))
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.CommitInfo{Votes: votes},
	})
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 3)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 97)),
		app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 1.82uband + 6*2% = 1.82uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(194, 2)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	// 14.406uband + 6*98% = 20.286uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(20286, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress).Rewards,
	)
	// 74.774uband
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(74774, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress).Rewards,
	)
}

func TestAllocateTokensWithDistrAllocateTokens(t *testing.T) {
	app, ctx := bandtesting.CreateTestApp(t, false)
	k := app.OracleKeeper

	ctx = ctx.WithBlockHeight(10) // Set block height to ensure distr's AllocateTokens gets called.
	votes := []abci.VoteInfo{{
		Validator:       abci.Validator{Address: bandtesting.Validators[0].PubKey.Address(), Power: 70},
		SignedLastBlock: true,
	}, {
		Validator:       abci.Validator{Address: bandtesting.Validators[1].PubKey.Address(), Power: 30},
		SignedLastBlock: true,
	}}

	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	distModule := app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)

	// Set collected fee to 100uband + 70% oracle reward proportion + disable minting inflation.
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 50)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
	)
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, feeCollector)
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	err = app.MintKeeper.SetParams(ctx, mintParams)
	require.NoError(t, err)
	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 70
	err = k.SetParams(ctx, params)
	require.NoError(t, err)
	// Set block proposer to Validators[1], who will receive 5% bonus.
	app.DistrKeeper.SetPreviousProposerConsAddr(ctx, bandtesting.Validators[1].Address.Bytes())
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	// Only Validators[0] active. After we call begin block:
	//   35uband = 70% go to oracle pool
	//     0.7uband (2%) go to community pool
	//     34.3uband go to Validators[0] (active)
	//   15uband = 30% go to distr pool
	//     0.3uband (2%) go to community pool
	//     2.25uband (15%) go to Validators[1] (proposer)
	//     12.45uband split among voters
	//        8.715uband (70%) go to Validators[0]
	//        3.735uband (30%) go to Validators[1]
	// In summary
	//   Community pool: 0.7 + 0.3 = 1
	//   Validators[0]: 34.3 + 8.715 = 43.015
	//   Validators[1]: 2.25 + 3.735 = 5.985
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(t, err)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.CommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.Coins{}, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
		app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(1)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(44590, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress).Rewards,
	)
	require.Equal(
		t,
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(4410, 3)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress).Rewards,
	)
}
