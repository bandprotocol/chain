package oracle_test

// TODO: Fix tests
import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
)

type ABCITestSuite struct {
	suite.Suite

	app *band.BandApp
}

func TestABCITestSuite(t *testing.T) {
	suite.Run(t, new(ABCITestSuite))
}

func (s *ABCITestSuite) SetupTest() {
	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})

	// Send all coins in the distribution module account.
	distModule := s.app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)
	err = s.app.BankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		distrtypes.ModuleName,
		bandtesting.Treasury.Address,
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	s.Require().NoError(err)
	err = s.app.DistrKeeper.FeePool.Set(ctx, distrtypes.InitialFeePool())
	s.Require().NoError(err)
}

func fromHex(hexStr string) []byte {
	res, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return res
}

func (s *ABCITestSuite) TestRollingSeed() {
	k := s.app.OracleKeeper
	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	require := s.Require()

	// Initially rolling seed should be all zeros.
	require.Equal(
		fromHex("0000000000000000000000000000000000000000000000000000000000000000"),
		k.GetRollingSeed(ctx),
	)
	// Every begin block, the rolling seed should get updated.
	_, err := s.app.BeginBlocker(ctx.WithHeaderInfo(header.Info{
		Hash: fromHex("0100000000000000000000000000000000000000000000000000000000000000"),
	}))
	require.Equal(
		fromHex("0000000000000000000000000000000000000000000000000000000000000001"),
		k.GetRollingSeed(ctx),
	)
	require.NoError(err)

	_, err = s.app.BeginBlocker(ctx.WithHeaderInfo(header.Info{
		Hash: fromHex("0200000000000000000000000000000000000000000000000000000000000000"),
	}))
	require.Equal(
		fromHex("0000000000000000000000000000000000000000000000000000000000000102"),
		k.GetRollingSeed(ctx),
	)
	require.NoError(err)

	_, err = s.app.BeginBlocker(ctx.WithHeaderInfo(header.Info{
		Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
	}))
	require.Equal(
		fromHex("00000000000000000000000000000000000000000000000000000000000102ff"),
		k.GetRollingSeed(ctx),
	)
	require.NoError(err)
}

func (s *ABCITestSuite) TestAllocateTokensCalledOnBeginBlock() {
	k := s.app.OracleKeeper
	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	require := s.Require()

	votes := []abci.VoteInfo{{
		Validator:   abci.Validator{Address: bandtesting.Validators[0].PubKey.Address(), Power: 70},
		BlockIdFlag: tmproto.BlockIDFlagCommit,
	}, {
		Validator:   abci.Validator{Address: bandtesting.Validators[1].PubKey.Address(), Power: 30},
		BlockIdFlag: tmproto.BlockIDFlagCommit,
	}}

	// Set collected fee to 10000uband + 70% oracle reward proportion + disable minting inflation.
	feeCollector := s.app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 10000)))
	require.NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 10000)),
	)
	require.NoError(err)
	s.app.AccountKeeper.SetAccount(ctx, feeCollector)

	distModule := s.app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)

	mintParams, err := s.app.MintKeeper.Params.Get(ctx)
	require.NoError(err)
	mintParams.InflationMin = math.LegacyZeroDec()
	mintParams.InflationMax = math.LegacyZeroDec()
	err = s.app.MintKeeper.Params.Set(ctx, mintParams)
	require.NoError(err)

	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 70
	err = k.SetParams(ctx, params)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 10000)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	// If there are no validators active, Calling begin block should be no-op.
	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[0].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 10000)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)

	// 1 validator active, begin block should take 70% of the fee. 2% of that goes to comm pool.
	err = k.Activate(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[0].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 3000)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 7000)),
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 10000*70%*2% = 140uband
	feePool, err := s.app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(140)}},
		feePool.CommunityPool,
	)
	// 0uband
	require.Empty(s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress))
	// 10000*70%*98% = 6860uband
	valOutReward, err := s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(6860)}},
		valOutReward.Rewards,
	)

	// 2 validators active now. 70% of the remaining fee pool will be split 3 ways (comm pool + val0 + val1).
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)

	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[0].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 900)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 9100)),
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 140uband + 3000*70%*2% = 182uband
	feePool, err = s.app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(182)}},
		feePool.CommunityPool,
	)
	// 3000*70%*98%*70% = 1440.6uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(14406, 1)}},
		valOutReward.Rewards,
	)
	// 68.6uband + 3000*70%*98%*30% = 7477.4uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(74774, 1)}},
		valOutReward.Rewards,
	)

	// 1 validator becomes inactive, and will not get reward this time.
	k.MissReport(ctx, bandtesting.Validators[1].ValAddress, bandtesting.ParseTime(100))
	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[0].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 270)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 9730)),
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 182uband + 630*2% = 194.6 but fund community pool function only distribute
	// to fee pool in integer amount so it will be 194uband.
	feePool, err = s.app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(194)}},
		feePool.CommunityPool,
	)
	// Since the validator is the only one active, it will get all the remaining fee pool.
	// 1440.6uband + 618 = 2058.6uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(20586, 1)}},
		valOutReward.Rewards,
	)
	// 7477.4uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(74774, 1)}},
		valOutReward.Rewards,
	)
}

func (s *ABCITestSuite) TestAllocateTokensToProposer() {
	k := s.app.OracleKeeper
	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	require := s.Require()

	votes := []abci.VoteInfo{
		{
			Validator:   abci.Validator{Address: bandtesting.Validators[0].PubKey.Address(), Power: 1},
			BlockIdFlag: tmproto.BlockIDFlagCommit,
		},
		{
			Validator:   abci.Validator{Address: bandtesting.Validators[1].PubKey.Address(), Power: 1},
			BlockIdFlag: tmproto.BlockIDFlagCommit,
		},
		{
			Validator:   abci.Validator{Address: bandtesting.Validators[2].PubKey.Address(), Power: 1},
			BlockIdFlag: tmproto.BlockIDFlagCommit,
		},
	}

	// Set collected fee to 100uband + 70% oracle reward proportion + disable minting inflation.
	feeCollector := s.app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 102)))
	require.NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 102)),
	)
	require.NoError(err)
	s.app.AccountKeeper.SetAccount(ctx, feeCollector)

	distModule := s.app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)

	mintParams, err := s.app.MintKeeper.Params.Get(ctx)
	require.NoError(err)
	mintParams.InflationMin = math.LegacyZeroDec()
	mintParams.InflationMax = math.LegacyZeroDec()
	err = s.app.MintKeeper.Params.Set(ctx, mintParams)
	require.NoError(err)

	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 100
	err = k.SetParams(ctx, params)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 102)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)

	// 3 validators active now. 70% of the remaining fee pool will be split 4 ways (comm pool + val0 + val1 + val2).
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	err = k.Activate(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	err = k.Activate(ctx, bandtesting.Validators[2].ValAddress)
	require.NoError(err)

	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[2].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 0)),
		s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()),
	)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 102)),
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	// 102*2% = 2.04uband but fund community pool function only distribute
	// to fee pool in integer amount so it will be 2uband.
	feePool, err := s.app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDec(2)}},
		feePool.CommunityPool,
	)
	// 100/3 = 33.3333333333333333uband
	valOutReward, err := s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(333333333333333333, 16)}},
		valOutReward.Rewards,
	)
	// 100/3 = 33.3333333333333333uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(333333333333333333, 16)}},
		valOutReward.Rewards,
	)
	// 100/3 = 33.3333333333333333uband
	// but the proposer get the remaining 0.0000000000000001uband
	// so the proposer will get 33.3333333333333334uband
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[2].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(333333333333333334, 16)}},
		valOutReward.Rewards,
	)
}

func (s *ABCITestSuite) TestAllocateTokensWithDistrAllocateTokens() {
	k := s.app.OracleKeeper
	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	require := s.Require()

	ctx = ctx.WithBlockHeight(10) // Set block height to ensure distr's AllocateTokens gets called.
	votes := []abci.VoteInfo{{
		Validator:   abci.Validator{Address: bandtesting.Validators[0].PubKey.Address(), Power: 70},
		BlockIdFlag: tmproto.BlockIDFlagCommit,
	}, {
		Validator:   abci.Validator{Address: bandtesting.Validators[1].PubKey.Address(), Power: 30},
		BlockIdFlag: tmproto.BlockIDFlagCommit,
	}}

	feeCollector := s.app.AccountKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	distModule := s.app.AccountKeeper.GetModuleAccount(ctx, distrtypes.ModuleName)

	// Set collected fee to 100uband + 70% oracle reward proportion + disable minting inflation.
	err := s.app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uband", 50)))
	require.NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(
		ctx,
		minttypes.ModuleName,
		authtypes.FeeCollectorName,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
	)
	require.NoError(err)

	s.app.AccountKeeper.SetAccount(ctx, feeCollector)
	mintParams, err := s.app.MintKeeper.Params.Get(ctx)
	require.NoError(err)
	mintParams.InflationMin = math.LegacyZeroDec()
	mintParams.InflationMax = math.LegacyZeroDec()
	err = s.app.MintKeeper.Params.Set(ctx, mintParams)
	require.NoError(err)

	params := k.GetParams(ctx)
	params.OracleRewardPercentage = 70
	err = k.SetParams(ctx, params)
	require.NoError(err)

	// Only Validators[0] active. After we call begin block:
	//   35uband = 70% go to oracle pool
	//     0.7uband (2%) -> 0uband go to community pool because oracle pool only distribute to fee pool in integer amount
	//     35uband go to Validators[0] (active)
	//   15uband = 30% go to distr pool
	//     0.3uband (2%) go to community pool
	//     14.7uband split among voters
	//        10.29uband (70%) go to Validators[0]
	//        4.41uband (30%) go to Validators[1]
	// In summary
	//   Community pool: 0.3 = 1 but oracle pool only distribute to fee pool in integer amount
	//     so it will be 0.3uband.
	//   Validators[0]: 35 + 10.29 = 45.29
	//   Validators[1]: 4.41
	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	// begin block with Validators[1] as proposer
	_, err = s.app.BeginBlocker(
		ctx.WithHeaderInfo(header.Info{Hash: fromHex("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")}).
			WithVoteInfos(votes).
			WithProposer(bandtesting.Validators[1].ValAddress.Bytes()),
	)
	require.NoError(err)
	require.Equal(sdk.Coins{}, s.app.BankKeeper.GetAllBalances(ctx, feeCollector.GetAddress()))
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
		s.app.BankKeeper.GetAllBalances(ctx, distModule.GetAddress()),
	)
	feePool, err := s.app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(3, 1)}},
		feePool.CommunityPool,
	)
	valOutReward, err := s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[0].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(4529, 2)}},
		valOutReward.Rewards,
	)
	valOutReward, err = s.app.DistrKeeper.GetValidatorOutstandingRewards(ctx, bandtesting.Validators[1].ValAddress)
	require.NoError(err)
	require.Equal(
		sdk.DecCoins{{Denom: "uband", Amount: math.LegacyNewDecWithPrec(441, 2)}},
		valOutReward.Rewards,
	)
}
