package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
)

var Coins1000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))

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
	// Set collected fee to 1000000uband and 50% tss reward proportion.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, Coins1000000uband)
	app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		Coins1000000uband,
	)
	app.AccountKeeper.SetAccount(ctx, feeCollector)

	params := k.GetParams(ctx)
	params.RewardPercentage = 50
	k.SetParams(ctx, params)

	return feeCollector
}

func (s *KeeperTestSuite) TestAllocateTokenNoActiveValidators() {
	app, ctx, k := testapp.CreateTestInput(false)
	feeCollector := SetupFeeCollector(app, ctx, app.TSSKeeper)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// No active oracle validators so nothing should happen.
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	s.Require().Empty(app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
}

func (s *KeeperTestSuite) TestAllocateTokensOneActive() {
	app, ctx, _ := testapp.CreateTestInput(false)
	k := app.TSSKeeper
	feeCollector := SetupFeeCollector(app, ctx, app.TSSKeeper)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// From 50% of fee, 1% should go to community pool, the rest goes to the only active validator.
	k.SetActive(ctx, testapp.Validators[1].Address)
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()),
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(10000)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	s.Require().Empty(app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress))
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(490000)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress).Rewards,
	)
	s.Require().Empty(app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[2].ValAddress))
}

func (s *KeeperTestSuite) TestAllocateTokensAllActive() {
	ctx, app, k := s.ctx, s.app, s.app.TSSKeeper
	feeCollector := SetupFeeCollector(app, ctx, k)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// From 50% of fee, 1% should go to community pool, the rest get split to validators.
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()),
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(10000)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(343000)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[0].ValAddress).Rewards,
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(98000)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[1].ValAddress).Rewards,
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(49000)}},
		app.DistrKeeper.GetValidatorOutstandingRewards(ctx, testapp.Validators[2].ValAddress).Rewards,
	)
}
