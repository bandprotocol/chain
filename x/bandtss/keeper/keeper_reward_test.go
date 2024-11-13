package keeper_test

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var Coins1000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))

func defaultVotes() []abci.VoteInfo {
	return []abci.VoteInfo{{
		Validator: abci.Validator{
			Address: bandtesting.Validators[0].PubKey.Address(),
			Power:   70,
		},
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}, {
		Validator: abci.Validator{
			Address: bandtesting.Validators[1].PubKey.Address(),
			Power:   20,
		},
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}, {
		Validator: abci.Validator{
			Address: bandtesting.Validators[2].PubKey.Address(),
			Power:   10,
		},
		BlockIdFlag: cmtproto.BlockIDFlagCommit,
	}}
}

func SetupFeeCollector(
	app *band.BandApp,
	ctx sdk.Context,
	k keeper.Keeper,
) (sdk.ModuleAccountI, error) {
	// Set collected fee to 1000000uband and 50% tss reward proportion.
	feeCollector := app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	if err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, Coins1000000uband); err != nil {
		return nil, err
	}

	// remove all coins from fee collector
	if err := app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		authtypes.FeeCollectorName,
		minttypes.ModuleName,
		app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	); err != nil {
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

func (s *AppTestSuite) TestAllocateTokensOneActive() {
	app, ctx := s.app, s.ctx
	tssKeeper, k := app.TSSKeeper, app.BandtssKeeper

	// setup fee collector
	feeCollector, err := SetupFeeCollector(app, ctx, k)
	s.Require().NoError(err)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))

	// create a new group
	groupCtx := s.SetupNewGroup(2, 1)
	k.SetCurrentGroup(ctx, types.NewCurrentGroup(groupCtx.GroupID, s.ctx.BlockTime()))

	alice := groupCtx.Accounts[0].Address
	bob := groupCtx.Accounts[1].Address
	// From 50% of fee, 1% should go to community pool, the rest goes to the only active validator.
	err = tssKeeper.EnqueueDEs(ctx, alice, []tsstypes.DE{
		{
			PubD: testutil.HexDecode("dddd"),
			PubE: testutil.HexDecode("eeee"),
		},
	})
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

	err = k.AllocateTokens(ctx)
	s.Require().NoError(err)

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
	feePool, err := app.DistrKeeper.FeePool.Get(ctx)
	s.Require().NoError(err)

	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(10000)}},
		feePool.CommunityPool,
	)
}

func (s *AppTestSuite) TestAllocateTokensAllActive() {
	ctx, app, k := s.ctx, s.app, s.app.BandtssKeeper

	feeCollector, err := SetupFeeCollector(app, ctx, k)
	s.Require().NoError(err)
	s.Require().Equal(Coins1000000uband, app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))

	groupCtx := s.SetupNewGroup(3, 2)
	k.SetCurrentGroup(ctx, types.NewCurrentGroup(groupCtx.GroupID, s.ctx.BlockTime()))

	for _, acc := range groupCtx.Accounts {
		deQueue := s.app.TSSKeeper.GetDEQueue(ctx, acc.Address)
		s.Require().Greater(deQueue.Tail, deQueue.Head)
	}

	// From 50% of fee, 1% should go to community pool, the rest get split to validators.)
	balancesBefore := make([]sdk.Coins, len(groupCtx.Accounts))
	for i, acc := range groupCtx.Accounts {
		balancesBefore[i] = app.BankKeeper.GetAllBalances(ctx, acc.Address)
	}

	err = k.AllocateTokens(ctx)
	s.Require().NoError(err)

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

	feePool, err := app.DistrKeeper.FeePool.Get(ctx)
	s.Require().NoError(err)
	s.Require().Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(10001)}},
		feePool.CommunityPool,
	)

	for i := range bandtesting.Validators {
		s.Require().Equal(
			sdk.NewCoins(sdk.NewInt64Coin("uband", 163333)),
			balancesAfter[i].Sub(balancesBefore[i]...),
			fmt.Sprintf("incorrect balance for validator %d", i),
		)
	}
}
