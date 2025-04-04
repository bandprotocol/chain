package bank_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

type AppTestSuite struct {
	suite.Suite

	app *band.BandApp
}

var (
	NoAbsentVotes = abci.CommitInfo{
		Votes: []abci.VoteInfo{
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[0].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[1].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[2].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.MissedValidator.PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
		},
	}
	AbsentVotes = abci.CommitInfo{
		Votes: []abci.VoteInfo{
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[0].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[1].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.Validators[2].PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagCommit,
			},
			{
				Validator:   abci.Validator{Address: bandtesting.MissedValidator.PubKey.Address().Bytes(), Power: 100},
				BlockIdFlag: cmtproto.BlockIDFlagAbsent,
			},
		},
	}
)

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupTest() {
	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	params, err := s.app.SlashingKeeper.GetParams(ctx)
	s.Require().NoError(err)

	// Set new sign window
	params.SignedBlocksWindow = 2
	params.MinSignedPerWindow = math.LegacyNewDecWithPrec(5, 1)

	// Add Missed validator
	res1 := s.app.AccountKeeper.GetAccount(ctx, bandtesting.MissedValidator.Address)
	s.Require().NotNil(res1)

	acc1Num := res1.GetAccountNumber()
	acc1Seq := res1.GetSequence()

	txConfig := moduletestutil.MakeTestTxConfig()

	err = s.app.SlashingKeeper.SetParams(ctx, params)
	s.Require().NoError(err)

	msg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(bandtesting.MissedValidator.Address).String(),
		bandtesting.MissedValidator.PubKey,
		sdk.NewInt64Coin("uband", 100000000),
		stakingtypes.Description{
			Moniker: "test",
		},
		stakingtypes.NewCommissionRates(
			math.LegacyOneDec(),
			math.LegacyOneDec(),
			math.LegacyOneDec(),
		),
		math.OneInt(),
	)
	s.Require().NoError(err)

	_, res, _, err := bandtesting.SignCheckDeliver(
		s.T(),
		txConfig,
		s.app.BaseApp,
		cmtproto.Header{Height: s.app.LastBlockHeight() + 1},
		[]sdk.Msg{msg},
		s.app.ChainID(),
		[]uint64{acc1Num},
		[]uint64{acc1Seq},
		true,
		true,
		bandtesting.MissedValidator.PrivKey,
	)
	s.Require().NotNil(res)
	s.Require().NoError(err)
}

func (s *AppTestSuite) checkCommunityPool(expected string) {
	ctx := s.app.NewUncachedContext(false, cmtproto.Header{})
	// Check community pool
	feePool, err := s.app.DistrKeeper.FeePool.Get(ctx)
	s.Require().NoError(err)

	dec, err := math.LegacyNewDecFromStr(expected)
	s.Require().NoError(err)

	s.Require().Equal(sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", dec)), feePool.CommunityPool)
}

func (s *AppTestSuite) TestNoAbsent() {
	// Pass 1 block no absent
	_, err := s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: NoAbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	s.checkCommunityPool("8326.44")

	// Pass 2 block no absent
	_, err = s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: NoAbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	s.checkCommunityPool("8408.88")

	// Pass 3 block no absent
	_, err = s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: NoAbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	s.checkCommunityPool("8491.32")
}

func (s *AppTestSuite) TestMissedValidatorAbsent() {
	// Pass 1 block absent nothing happen
	_, err := s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: AbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	s.checkCommunityPool("8326.44")

	// Pass 2 block absent missed validator not slash yet due to not pass min height
	_, err = s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: AbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	ctx := s.app.NewUncachedContext(false, cmtproto.Header{})
	missVal, err := s.app.StakingKeeper.GetValidator(ctx, bandtesting.MissedValidator.ValAddress)
	s.Require().NoError(err)
	s.Require().False(missVal.IsJailed())

	s.checkCommunityPool("8408.88")

	// Pass 3 block still miss should be slashed
	_, err = s.app.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, DecidedLastCommit: AbsentVotes},
	)
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)

	ctx = s.app.NewUncachedContext(false, cmtproto.Header{})
	missVal, err = s.app.StakingKeeper.GetValidator(ctx, bandtesting.MissedValidator.ValAddress)
	s.Require().NoError(err)
	s.Require().True(missVal.IsJailed())

	// Community pool should increase 1% of validator power(100 band) == 1 band == 1000000uband
	s.checkCommunityPool("1008491.32")
}
