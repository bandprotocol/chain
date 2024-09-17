package keeper_test

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var Coins1000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))

func defaultVotes() []abci.VoteInfo {
	return []abci.VoteInfo{{
		Validator: abci.Validator{
			Address: bandtesting.Validators[0].PubKey.Address(),
			Power:   70,
		},
		SignedLastBlock: true,
	}, {
		Validator: abci.Validator{
			Address: bandtesting.Validators[1].PubKey.Address(),
			Power:   20,
		},
		SignedLastBlock: true,
	}, {
		Validator: abci.Validator{
			Address: bandtesting.Validators[2].PubKey.Address(),
			Power:   10,
		},
		SignedLastBlock: true,
	}}
}

func SetupFeeCollector(
	app *bandtesting.TestingApp,
	ctx sdk.Context,
	k keeper.Keeper,
) (authtypes.ModuleAccountI, error) {
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

func (s *AppTestSuite) TestAllocateTokenNoActiveValidators() {
	app, ctx := bandtesting.CreateTestApp(s.T(), false)
	feeCollector, err := SetupFeeCollector(app, ctx, *app.BandtssKeeper)
	s.Require().NoError(err)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	// No active tss validators so nothing should happen.
	app.OracleKeeper.AllocateTokens(ctx, defaultVotes())

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	s.Require().Empty(app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()))
}

func (s *AppTestSuite) TestAllocateTokensOneActive() {
	app, ctx := bandtesting.CreateTestApp(s.T(), false)
	tssKeeper, k := app.TSSKeeper, app.BandtssKeeper
	feeCollector, err := SetupFeeCollector(app, ctx, *k)
	s.Require().NoError(err)

	groupCtx := s.SetupNewGroup(2, 1)
	k.SetCurrentGroupID(ctx, groupCtx.GroupID)

	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	alice := groupCtx.Accounts[0].Address
	bob := groupCtx.Accounts[1].Address
	// From 50% of fee, 1% should go to community pool, the rest goes to the only active validator.
	err = tssKeeper.HandleSetDEs(ctx, alice, []tsstypes.DE{
		{
			PubD: testutil.HexDecode("dddd"),
			PubE: testutil.HexDecode("eeee"),
		},
	})
	s.Require().NoError(err)

	err = k.AddMember(ctx, alice, groupCtx.GroupID)
	s.Require().NoError(err)

	err = k.AddMember(ctx, bob, groupCtx.GroupID)
	s.Require().NoError(err)

	tssKeeper.SetMember(ctx, tsstypes.Member{
		ID:       tss.MemberID(1),
		GroupID:  1,
		Address:  alice.String(),
		IsActive: true,
	})
	tssKeeper.SetMember(ctx, tsstypes.Member{
		ID:       tss.MemberID(2),
		GroupID:  1,
		Address:  bob.String(),
		IsActive: false,
	})

	aliceBalanceBefore := app.BankKeeper.GetAllBalances(ctx, alice)
	bobBalanceBefore := app.BankKeeper.GetAllBalances(ctx, bob)
	k.AllocateTokens(ctx)
	aliceBalanceAfter := app.BankKeeper.GetAllBalances(ctx, alice)
	bobBalanceAfter := app.BankKeeper.GetAllBalances(ctx, bob)

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 490000)),
		aliceBalanceAfter.Sub(aliceBalanceBefore...),
	)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 0)),
		bobBalanceAfter.Sub(bobBalanceBefore...),
	)

	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 10000)),
		app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()),
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(10000)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)
}

func (s *AppTestSuite) TestAllocateTokensAllActive() {
	ctx, app, k := s.ctx, s.app, s.app.BandtssKeeper

	feeCollector, err := SetupFeeCollector(app, ctx, *k)
	s.Require().NoError(err)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))

	groupCtx := s.SetupNewGroup(3, 2)
	k.SetCurrentGroupID(ctx, groupCtx.GroupID)

	for _, acc := range groupCtx.Accounts {
		deQueue := s.app.TSSKeeper.GetDEQueue(ctx, acc.Address)
		s.Require().Greater(deQueue.Tail, deQueue.Head)
	}

	// From 50% of fee, 1% should go to community pool, the rest get split to validators.)
	balancesBefore := make([]sdk.Coins, len(groupCtx.Accounts))
	for i, acc := range groupCtx.Accounts {
		balancesBefore[i] = app.BankKeeper.GetAllBalances(ctx, acc.Address)
	}

	k.AllocateTokens(ctx)

	balancesAfter := make([]sdk.Coins, len(groupCtx.Accounts))
	for i, acc := range groupCtx.Accounts {
		balancesAfter[i] = app.BankKeeper.GetAllBalances(ctx, acc.Address)
	}

	distAccount := app.AccountKeeper.GetModuleAccount(ctx, disttypes.ModuleName)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 500000)),
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	s.Require().Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 10001)),
		app.BankKeeper.GetAllBalances(ctx, distAccount.GetAddress()),
	)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(10001)}},
		app.DistrKeeper.GetFeePool(ctx).CommunityPool,
	)

	for i := range bandtesting.Validators {
		s.Require().Equal(
			sdk.NewCoins(sdk.NewInt64Coin("uband", 163333)),
			balancesAfter[i].Sub(balancesBefore[i]...),
			fmt.Sprintf("incorrect balance for validator %d", i),
		)
	}
}