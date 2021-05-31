package oracle_test

import (
	"encoding/hex"
	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func fromHex(hexStr string) []byte {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return res
}

func TestRollingSeedCorrect(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(false, true)
	// Initially rolling seed should be all zeros.
	require.Equal(t, fromHex("0000000000000000000000000000000000000000000000000000000000000000"), k.GetRollingSeed(ctx))
	// Every begin block, the rolling seed should get updated.
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0100000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(t, fromHex("0000000000000000000000000000000000000000000000000000000000000001"), k.GetRollingSeed(ctx))
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("0200000000000000000000000000000000000000000000000000000000000000"),
	})
	require.Equal(t, fromHex("0000000000000000000000000000000000000000000000000000000000000102"), k.GetRollingSeed(ctx))
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	})
	require.Equal(t, fromHex("00000000000000000000000000000000000000000000000000000000000102ff"), k.GetRollingSeed(ctx))
}

func TestAllocateTokensCalledOnBeginBlock(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(false, false)

	app.SlashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)

	votes := []abci.VoteInfo{{
		Validator:       abci.Validator{Address: testapp.Validators[0].PubKey.Address(), Power: 70},
		SignedLastBlock: true,
	}, {
		Validator:       abci.Validator{Address: testapp.Validators[1].PubKey.Address(), Power: 30},
		SignedLastBlock: true,
	}}
	// Set collected fee to 100loki + 70% oracle reward proportion + disable minting inflation.
	// NOTE: we intentionally keep ctx.BlockHeight = 0, so distr's AllocateTokens doesn't get called.

	app.BankKeeper.SetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), sdk.NewInt64Coin("loki", 100))
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	app.AccountKeeper.SetAccount(ctx, feeCollector)
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	app.MintKeeper.SetParams(ctx, mintParams)
	k.SetParamUint64(ctx, types.KeyOracleRewardPercentage, 70)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 100)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	// If there are no validators active, Calling begin block should be no-op.
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.LastCommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 100)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	// 1 validator active, begin block should take 70% of the fee. 2% of that goes to comm pool.
	k.Activate(ctx, testapp.Validators[1].ValAddress)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.LastCommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 30)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 70)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)))
	// 100*70%*2% = 1.4loki
	require.Equal(t, sdk.DecCoins{{Denom: "loki", Amount: sdk.NewDecWithPrec(14, 1)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	// 0loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins(nil)}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	// 100*70%*98% = 68.6loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(686, 1)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress))
	// 2 validators active now. 70% of the remaining fee pool will be split 3 ways (comm pool + val1 + val2).
	k.Activate(ctx, testapp.Validators[0].ValAddress)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.LastCommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 9)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 91)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)))
	// 1.4loki + 30*70%*2% = 1.82loki
	require.Equal(t, sdk.DecCoins{{Denom: "loki", Amount: sdk.NewDecWithPrec(182, 2)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	// 30*70%*98%*70% = 14.406loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(14406, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	// 68.6loki + 30*70%*98%*30% = 74.774loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(74774, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress))
	// 1 validator becomes in active, and will not get reward this time.
	k.MissReport(ctx, testapp.Validators[1].ValAddress, testapp.ParseTime(100))
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.LastCommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 3)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 97)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)))
	// 1.82loki + 6*2% = 1.82loki
	require.Equal(t, sdk.DecCoins{{Denom: "loki", Amount: sdk.NewDecWithPrec(194, 2)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	// 14.406loki + 6*98% = 20.286loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(20286, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	// 74.774loki
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(74774, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress))
}

func TestAllocateTokensWithDistrAllocateTokens(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(false)
	ctx = ctx.WithBlockHeight(10) // Set block height to ensure distr's AllocateTokens gets called.
	votes := []abci.VoteInfo{{
		Validator:       abci.Validator{Address: testapp.Validators[0].PubKey.Address(), Power: 70},
		SignedLastBlock: true,
	}, {
		Validator:       abci.Validator{Address: testapp.Validators[1].PubKey.Address(), Power: 30},
		SignedLastBlock: true,
	}}
	// Set collected fee to 100loki + 70% oracle reward proportion + disable minting inflation.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	app.BankKeeper.SetBalance(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName), sdk.NewInt64Coin("loki", 50))
	app.AccountKeeper.SetAccount(ctx, feeCollector)
	mintParams := app.MintKeeper.GetParams(ctx)
	mintParams.InflationMin = sdk.ZeroDec()
	mintParams.InflationMax = sdk.ZeroDec()
	app.MintKeeper.SetParams(ctx, mintParams)
	k.SetParamUint64(ctx, types.KeyOracleRewardPercentage, 70)
	// Set block proposer to Validator2, who will receive 5% bonus.
	app.DistrKeeper.SetPreviousProposerConsAddr(ctx, testapp.Validators[1].Address.Bytes())
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 50)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	// Only validator1 active. After we call begin block:
	//   35loki = 70% go to oracle pool
	//     0.7loki (2%) go to community pool
	//     34.3loki go to validator1 (active)
	//   15loki = 30% go to distr pool
	//     0.3loki (2%) go to community pool
	//     2.25loki (15%) go to validator2 (proposer)
	//     12.45loki split among voters
	//        8.715loki (70%) go to validator1
	//        3.735loki (30%) go to validator2
	// In summary
	//   Community pool: 0.7 + 0.3 = 1
	//   Validator1: 34.3 + 8.715 = 43.015
	//   Validator2: 2.25 + 3.735 = 5.985
	k.Activate(ctx, testapp.Validators[0].ValAddress)
	app.BeginBlocker(ctx, abci.RequestBeginBlock{
		Hash:           fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
		LastCommitInfo: abci.LastCommitInfo{Votes: votes},
	})
	require.Equal(t, sdk.Coins(nil), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("loki", 50)), app.BankKeeper.GetAllBalances(ctx, app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)))
	require.Equal(t, sdk.DecCoins{{Denom: "loki", Amount: sdk.NewDec(1)}}, app.DistrKeeper.GetFeePool(ctx).CommunityPool)
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(43015, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	require.Equal(t, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.NewDecCoins(sdk.NewDecCoinFromDec("loki", sdk.NewDecWithPrec(5985, 3)))}, app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress))
}
