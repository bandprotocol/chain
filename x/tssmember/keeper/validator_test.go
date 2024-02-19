package keeper_test

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tssmember/keeper"
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

func SetupFeeCollector(app *testapp.TestingApp, ctx sdk.Context, k keeper.Keeper) (authtypes.ModuleAccountI, error) {
	// Set collected fee to 1000000uband and 50% tss reward proportion.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, Coins1000000uband); err != nil {
		return nil, err
	}

	if err := app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		Coins1000000uband,
	); err != nil {
		return nil, err
	}
	app.AccountKeeper.SetAccount(ctx, feeCollector)

	params := k.GetParams(ctx)
	params.RewardPercentage = 50
	if err := k.SetParams(ctx, params); err != nil {
		return nil, err
	}

	return feeCollector, nil
}

func (s *KeeperTestSuite) TestAllocateTokenNoActiveValidators() {
	app, ctx, k := testapp.CreateTestInput(false)
	feeCollector, err := SetupFeeCollector(app, ctx, app.TSSMemberKeeper)
	s.Require().NoError(err)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// No active tss validators so nothing should happen.
	k.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	s.Require().Empty(app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
}

func (s *KeeperTestSuite) TestAllocateTokensOneActive() {
	app, ctx, _ := testapp.CreateTestInput(false)
	tssKeeper, k := app.TSSKeeper, app.TSSMemberKeeper
	feeCollector, err := SetupFeeCollector(app, ctx, k)
	s.Require().NoError(err)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// From 50% of fee, 1% should go to community pool, the rest goes to the only active validator.
	err = tssKeeper.HandleSetDEs(ctx, testapp.Validators[1].Address, []tsstypes.DE{
		{
			PubD: testutil.HexDecode("dddd"),
			PubE: testutil.HexDecode("eeee"),
		},
	})
	s.Require().NoError(err)

	err = tssKeeper.SetActiveStatus(ctx, testapp.Validators[1].Address)
	s.Require().NoError(err)

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
	ctx, app, k := s.ctx, s.app, s.app.TSSMemberKeeper

	feeCollector, err := SetupFeeCollector(app, ctx, k)
	s.Require().NoError(err)

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

func (s *KeeperTestSuite) TestHandleInactiveValidators() {
	ctx, k, tssKeeper := s.ctx, s.app.TSSMemberKeeper, s.app.TSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := testapp.Validators[0].Address

	status := tsstypes.Status{
		Status:     tsstypes.MEMBER_STATUS_ACTIVE,
		Address:    address.String(),
		Since:      time.Time{},
		LastActive: time.Time{},
	}
	tssKeeper.SetMemberStatus(ctx, status)
	ctx = ctx.WithBlockTime(time.Now())

	k.HandleInactiveValidators(ctx)

	status = tssKeeper.GetStatus(ctx, address)
	s.Require().Equal(tsstypes.MEMBER_STATUS_INACTIVE, status.Status)
}
