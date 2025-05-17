package band_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/app/mempool"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstestutils "github.com/bandprotocol/chain/v3/x/tss/testutil"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

type AppTestSuite struct {
	suite.Suite

	app      *band.BandApp
	ctx      sdk.Context
	txConfig client.TxConfig

	valAccWithNumSeq      bandtesting.AccountWithNumSeq
	feederAccWithNumSeq   bandtesting.AccountWithNumSeq
	reporterAccWithNumSeq bandtesting.AccountWithNumSeq

	tssAccountsWithNumSeq   []bandtesting.AccountWithNumSeq
	tssGranteeAccWithNumSeq bandtesting.AccountWithNumSeq
	tssGroupCtx             *tsstestutils.GroupContext
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupTest() {
	dir := testutil.GetTempDir(s.T())
	s.app = bandtesting.SetupWithCustomHome(false, dir)
	s.txConfig = moduletestutil.MakeTestTxConfig()
	ctx := s.app.NewUncachedContext(false, cmtproto.Header{})

	// Activate validators
	for _, v := range bandtesting.Validators {
		err := s.app.OracleKeeper.Activate(ctx, v.ValAddress)
		s.Require().NoError(err)
	}

	s.app.OracleKeeper.SetRequest(
		ctx,
		1,
		oracletypes.NewRequest(
			1,
			[]byte("calldata"),
			[]sdk.ValAddress{
				bandtesting.Validators[0].ValAddress,
			},
			1,
			1,
			ctx.BlockTime(),
			"",
			[]oracletypes.RawRequest{
				oracletypes.NewRawRequest(1, 1, []byte("test")),
				oracletypes.NewRawRequest(2, 1, []byte("test")),
				oracletypes.NewRawRequest(3, 1, []byte("test")),
			},
			nil,
			0,
			0,
			bandtesting.FeePayer.Address.String(),
			bandtesting.Coins100band,
		),
	)

	// Set authorization for feeders
	feederAcc := bandtesting.Alice.Address
	err := s.app.AuthzKeeper.SaveGrant(
		ctx,
		feederAcc,
		bandtesting.Validators[0].Address,
		authz.NewGenericAuthorization(
			sdk.MsgTypeURL(&feedstypes.MsgSubmitSignalPrices{}),
		),
		nil,
	)
	s.Require().NoError(err)

	// Set authorization for reporter
	reporterAcc := bandtesting.Bob.Address
	err = s.app.AuthzKeeper.SaveGrant(
		ctx,
		reporterAcc,
		bandtesting.Validators[0].Address,
		authz.NewGenericAuthorization(
			sdk.MsgTypeURL(&oracletypes.MsgReportData{}),
		),
		nil,
	)
	s.Require().NoError(err)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)

	_, err = s.app.Commit()
	s.Require().NoError(err)

	// Get account numbers and sequences
	info := s.app.AccountKeeper.GetAccount(ctx, bandtesting.Validators[0].Address)
	s.valAccWithNumSeq = bandtesting.AccountWithNumSeq{
		Account: bandtesting.Validators[0],
		Num:     info.GetAccountNumber(),
		Seq:     info.GetSequence(),
	}

	info = s.app.AccountKeeper.GetAccount(ctx, feederAcc)
	s.feederAccWithNumSeq = bandtesting.AccountWithNumSeq{
		Account: bandtesting.Alice,
		Num:     info.GetAccountNumber(),
		Seq:     info.GetSequence(),
	}

	info = s.app.AccountKeeper.GetAccount(ctx, reporterAcc)
	s.reporterAccWithNumSeq = bandtesting.AccountWithNumSeq{
		Account: bandtesting.Bob,
		Num:     info.GetAccountNumber(),
		Seq:     info.GetSequence(),
	}

	s.app.FeedsKeeper.SetCurrentFeeds(ctx, GenFeeds(300))

	// Setup an account for TSS group
	groupSize := uint64(3)
	threshold := uint64(3)
	tssAccounts := []bandtesting.Account{
		bandtesting.Alice,
		bandtesting.Bob,
		bandtesting.Carol,
	}

	tssMembers := make([]string, groupSize)
	for i := range tssAccounts {
		tssMembers[i] = tssAccounts[i].Address.String()
	}

	s.tssGranteeAccWithNumSeq = bandtesting.AccountWithNumSeq{
		Account: bandtesting.Validators[0],
		Num:     s.app.AccountKeeper.GetAccount(ctx, bandtesting.Validators[0].Address).GetAccountNumber(),
		Seq:     s.app.AccountKeeper.GetAccount(ctx, bandtesting.Validators[0].Address).GetSequence(),
	}

	tssGrantMsgTypes := []string{
		sdk.MsgTypeURL(&tsstypes.MsgSubmitDKGRound1{}),
		sdk.MsgTypeURL(&tsstypes.MsgSubmitDKGRound2{}),
		sdk.MsgTypeURL(&tsstypes.MsgConfirm{}),
		sdk.MsgTypeURL(&tsstypes.MsgComplain{}),
		sdk.MsgTypeURL(&tsstypes.MsgSubmitDEs{}),
		sdk.MsgTypeURL(&tsstypes.MsgSubmitSignature{}),
	}

	s.tssAccountsWithNumSeq = make([]bandtesting.AccountWithNumSeq, 0)

	for _, account := range tssAccounts {
		info := s.app.AccountKeeper.GetAccount(ctx, account.Address)
		s.tssAccountsWithNumSeq = append(s.tssAccountsWithNumSeq, bandtesting.AccountWithNumSeq{
			Account: account,
			Num:     info.GetAccountNumber(),
			Seq:     info.GetSequence(),
		})

		for _, msgType := range tssGrantMsgTypes {
			err := s.app.AuthzKeeper.SaveGrant(
				ctx,
				s.tssGranteeAccWithNumSeq.Address,
				account.Address,
				authz.NewGenericAuthorization(msgType),
				nil,
			)
			s.Require().NoError(err)
		}
	}

	// Create BandTSS Group
	bandtssMsgServer := bandtsskeeper.NewMsgServerImpl(s.app.BandtssKeeper)
	_, err = bandtssMsgServer.TransitionGroup(ctx, &bandtsstypes.MsgTransitionGroup{
		Members:   tssMembers,
		Threshold: threshold,
		ExecTime:  ctx.BlockTime().AddDate(0, 0, 2),
		Authority: s.app.BandtssKeeper.GetAuthority(),
	})
	s.Require().NoError(err)

	transition, found := s.app.BandtssKeeper.GetGroupTransition(ctx)
	s.Require().True(found)

	groupID := transition.IncomingGroupID
	dkgContext, err := s.app.TSSKeeper.GetDKGContext(ctx, groupID)
	s.Require().NoError(err)

	s.tssGroupCtx, err = tsstestutils.NewGroupContext(tssAccounts, groupID, threshold, dkgContext)
	s.Require().NoError(err)

	consensusParams := *bandtesting.DefaultConsensusParams
	consensusParams.Block.MaxGas = 50000000
	err = s.app.ConsensusParamsKeeper.ParamsStore.Set(ctx, consensusParams)
	s.Require().NoError(err)

	ctx = ctx.WithConsensusParams(consensusParams)
	s.ctx = ctx
}

// -----------------------------------------------
// FeedsLane tests
// -----------------------------------------------

// TestFeedsLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestFeedsLaneZeroGas() {
	msg := GenMsgSubmitSignalPrices(
		&s.valAccWithNumSeq,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)

	s.checkLaneWithZeroGas(msg, s.valAccWithNumSeq)
}

// TestFeedsLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestFeedsLaneExactGas() {
	msg := GenMsgSubmitSignalPrices(
		&s.valAccWithNumSeq,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)

	s.checkLaneWithExactGas(msg, s.valAccWithNumSeq, "feedsLane", 1000000, sdk.Coins{})
}

// TestFeedsLaneExceedGas tests that transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestFeedsLaneExceedGas() {
	msg := GenMsgSubmitSignalPrices(
		&s.valAccWithNumSeq,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)

	s.checkLaneWithExceedGas(msg, s.valAccWithNumSeq, "feedsLane", 1000001, sdk.Coins{})
}

// TestFeedsLaneWrappedMsgExactGas tests that transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestFeedsLaneWrappedMsgExactGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.feederAccWithNumSeq
	msg := GenMsgSubmitSignalPrices(
		&txOwner,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)
	s.checkLaneWrappedMsgWithExactGas(msg, sender, "feedsLane", 1000000, sdk.Coins{})
}

// TestFeedsLaneWrappedMsgExceedGas tests that transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestFeedsLaneWrappedMsgExceedGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.feederAccWithNumSeq
	msg := GenMsgSubmitSignalPrices(
		&txOwner,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "feedsLane", 1000001, sdk.Coins{})
}

// -----------------------------------------------
// TSSLane tests
// -----------------------------------------------

// TestTSSLaneMsgSubmitDKGRound1ZeroGas tests that MsgSubmitDKGRound1 transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound1ZeroGas() {
	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound1(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.Round1Infos[0],
	)

	s.checkLaneWithZeroGas(msg, sender)
}

// TestTSSLaneMsgSubmitDKGRound2ZeroGas tests that MsgSubmitDKGRound2 transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound2ZeroGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound2(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.EncryptedSecretShares[0],
	)

	s.checkLaneWithZeroGas(msg, sender)
}

// TestTSSLaneMsgConfirmZeroGas tests that MsgConfirm transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgConfirmZeroGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgConfirm(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.OwnPubKeySigs[0],
	)

	s.checkLaneWithZeroGas(msg, sender)
}

// TestTSSLaneMsgComplainZeroGas tests that MsgComplain transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgComplainZeroGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	// Note: complainantID and respondentID are memberID which is 1-indexed.
	msg, err := genMsgComplain(
		&s.tssAccountsWithNumSeq[0],
		s.tssGroupCtx.GroupID,
		s.tssGroupCtx.Round1Infos,
		1,
		2,
	)
	s.Require().NoError(err)

	s.checkLaneWithZeroGas(msg, s.tssAccountsWithNumSeq[0])
}

// TestTSSLaneMsgSubmitDEZeroGas tests that MsgSubmitDE transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDEZeroGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	de := tsstestutils.GenerateDE(s.tssGroupCtx.Secrets[0])
	msg := genMsgSubmitDE(
		&s.tssAccountsWithNumSeq[0],
		[]tsstestutils.DEWithPrivateNonce{de},
	)

	s.checkLaneWithZeroGas(msg, s.tssAccountsWithNumSeq[0])
}

// TestTSSLaneMsgSubmitSignatureZeroGas tests that MsgSubmitSignature transactions with zero gas are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitSignatureZeroGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.FillDEs(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	originator := tsstypes.NewDirectOriginator(
		"targetChain",
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"test",
	)

	signingID, err := s.app.TSSKeeper.RequestSigning(
		s.ctx,
		s.tssGroupCtx.GroupID,
		&originator,
		tsstypes.NewTextSignatureOrder([]byte("msg")),
	)
	s.Require().NoError(err)

	signing, err := s.app.TSSKeeper.GetSigning(s.ctx, signingID)
	s.Require().NoError(err)

	sa, err := s.app.TSSKeeper.GetSigningAttempt(s.ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)

	de, err := s.tssGroupCtx.PopDE(
		0,
		sa.AssignedMembers[0].PubD,
		sa.AssignedMembers[0].PubE,
	)
	s.Require().NoError(err)

	assignedMembers := tsstypes.AssignedMembers(sa.AssignedMembers)
	signature, err := tsstestutils.GenerateSignature(
		signing,
		assignedMembers,
		tss.MemberID(1),
		de,
		s.tssGroupCtx.OwnPrivKeys[0],
	)
	s.Require().NoError(err)

	msg := genMsgSubmitSignature(
		&s.tssAccountsWithNumSeq[0],
		signingID,
		tss.MemberID(1),
		signature,
	)

	s.checkLaneWithZeroGas(msg, s.tssAccountsWithNumSeq[0])
}

// TestTSSLaneMsgSubmitDKGRound1ExactGas tests that MsgSubmitDKGRound1 transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound1ExactGas() {
	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound1(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.Round1Infos[0],
	)

	s.checkLaneWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgSubmitDKGRound2ExactGas tests that MsgSubmitDKGRound2 transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound2ExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound2(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.EncryptedSecretShares[0],
	)

	s.checkLaneWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgConfirmExactGas tests that MsgConfirm transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgConfirmExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgConfirm(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.OwnPubKeySigs[0],
	)

	s.checkLaneWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgComplainExactGas tests that MsgComplain transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgComplainExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	// Note: complainantID and respondentID are memberID which is 1-indexed.
	msg, err := genMsgComplain(
		&s.tssAccountsWithNumSeq[0],
		s.tssGroupCtx.GroupID,
		s.tssGroupCtx.Round1Infos,
		1,
		2,
	)
	s.Require().NoError(err)

	s.checkLaneWithExactGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgSubmitDEExactGas tests that MsgSubmitDE transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgSubmitDEExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	de := tsstestutils.GenerateDE(s.tssGroupCtx.Secrets[0])
	msg := genMsgSubmitDE(
		&s.tssAccountsWithNumSeq[0],
		[]tsstestutils.DEWithPrivateNonce{de},
	)

	s.checkLaneWithExactGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgSubmitSignatureExactGas tests that MsgSubmitSignature transactions with exact gas are accepted
func (s *AppTestSuite) TestTSSLaneMsgSubmitSignatureExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.FillDEs(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	originator := tsstypes.NewDirectOriginator(
		"targetChain",
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"test",
	)

	signingID, err := s.app.TSSKeeper.RequestSigning(
		s.ctx,
		s.tssGroupCtx.GroupID,
		&originator,
		tsstypes.NewTextSignatureOrder([]byte("msg")),
	)
	s.Require().NoError(err)

	signing, err := s.app.TSSKeeper.GetSigning(s.ctx, signingID)
	s.Require().NoError(err)

	sa, err := s.app.TSSKeeper.GetSigningAttempt(s.ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)

	de, err := s.tssGroupCtx.PopDE(
		0,
		sa.AssignedMembers[0].PubD,
		sa.AssignedMembers[0].PubE,
	)
	s.Require().NoError(err)

	assignedMembers := tsstypes.AssignedMembers(sa.AssignedMembers)
	signature, err := tsstestutils.GenerateSignature(
		signing,
		assignedMembers,
		tss.MemberID(1),
		de,
		s.tssGroupCtx.OwnPrivKeys[0],
	)
	s.Require().NoError(err)

	msg := genMsgSubmitSignature(
		&s.tssAccountsWithNumSeq[0],
		signingID,
		tss.MemberID(1),
		signature,
	)

	s.checkLaneWithExactGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneMsgSubmitDKGRound1ExceedGas tests that MsgSubmitDKGRound1 transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound1ExceedGas() {
	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound1(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.Round1Infos[0],
	)

	s.checkLaneWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneMsgSubmitDKGRound2ExceedGas tests that MsgSubmitDKGRound2 transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDKGRound2ExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgSubmitDKGRound2(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.EncryptedSecretShares[0],
	)

	s.checkLaneWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneMsgConfirmExceedGas tests that MsgConfirm transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgConfirmExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	sender := s.tssAccountsWithNumSeq[0]
	msg := genMsgConfirm(
		&sender,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.OwnPubKeySigs[0],
	)

	s.checkLaneWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneMsgComplainExceedGas tests that MsgComplain transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgComplainExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	// Note: complainantID and respondentID are memberID which is 1-indexed.
	msg, err := genMsgComplain(
		&s.tssAccountsWithNumSeq[0],
		s.tssGroupCtx.GroupID,
		s.tssGroupCtx.Round1Infos,
		1,
		2,
	)
	s.Require().NoError(err)

	s.checkLaneWithExceedGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneMsgSubmitDEExceedGas tests that MsgSubmitDE transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitDEExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	de := tsstestutils.GenerateDE(s.tssGroupCtx.Secrets[0])
	msg := genMsgSubmitDE(
		&s.tssAccountsWithNumSeq[0],
		[]tsstestutils.DEWithPrivateNonce{de},
	)

	s.checkLaneWithExceedGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneMsgSubmitSignatureExceedGas tests that MsgSubmitSignature transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestTSSLaneMsgSubmitSignatureExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.FillDEs(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	originator := tsstypes.NewDirectOriginator(
		"targetChain",
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"test",
	)

	signingID, err := s.app.TSSKeeper.RequestSigning(
		s.ctx,
		s.tssGroupCtx.GroupID,
		&originator,
		tsstypes.NewTextSignatureOrder([]byte("msg")),
	)
	s.Require().NoError(err)

	signing, err := s.app.TSSKeeper.GetSigning(s.ctx, signingID)
	s.Require().NoError(err)

	sa, err := s.app.TSSKeeper.GetSigningAttempt(s.ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)

	de, err := s.tssGroupCtx.PopDE(
		0,
		sa.AssignedMembers[0].PubD,
		sa.AssignedMembers[0].PubE,
	)
	s.Require().NoError(err)

	assignedMembers := tsstypes.AssignedMembers(sa.AssignedMembers)
	signature, err := tsstestutils.GenerateSignature(
		signing,
		assignedMembers,
		tss.MemberID(1),
		de,
		s.tssGroupCtx.OwnPrivKeys[0],
	)
	s.Require().NoError(err)

	msg := genMsgSubmitSignature(
		&s.tssAccountsWithNumSeq[0],
		signingID,
		tss.MemberID(1),
		signature,
	)

	s.checkLaneWithExceedGas(msg, s.tssAccountsWithNumSeq[0], "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDKGRound1ExactGas tests that MsgSubmitDKGRound1 transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDKGRound1ExactGas() {
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDKGRound1(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.Round1Infos[0],
	)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDKGRound2ExcactGas tests that MsgSubmitDKGRound2 transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDKGRound2ExcactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDKGRound2(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.EncryptedSecretShares[0],
	)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgConfirmExactGas tests that MsgConfirm transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgConfirmExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgConfirm(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.OwnPubKeySigs[0],
	)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgComplainExactGas tests that MsgComplain transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgComplainExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	// Note: complainantID and respondentID are memberID which is 1-indexed.
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg, err := genMsgComplain(
		&txOwner,
		s.tssGroupCtx.GroupID,
		s.tssGroupCtx.Round1Infos,
		1,
		2,
	)
	s.Require().NoError(err)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDEExactGas tests that MsgSubmitDE transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDEExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	de := tsstestutils.GenerateDE(s.tssGroupCtx.Secrets[0])
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDE(
		&txOwner,
		[]tsstestutils.DEWithPrivateNonce{de},
	)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitSignatureExactGas tests that MsgSubmitSignature transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitSignatureExactGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.FillDEs(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	originator := tsstypes.NewDirectOriginator(
		"targetChain",
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"test",
	)

	signingID, err := s.app.TSSKeeper.RequestSigning(
		s.ctx,
		s.tssGroupCtx.GroupID,
		&originator,
		tsstypes.NewTextSignatureOrder([]byte("msg")),
	)
	s.Require().NoError(err)

	signing, err := s.app.TSSKeeper.GetSigning(s.ctx, signingID)
	s.Require().NoError(err)

	sa, err := s.app.TSSKeeper.GetSigningAttempt(s.ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)

	de, err := s.tssGroupCtx.PopDE(
		0,
		sa.AssignedMembers[0].PubD,
		sa.AssignedMembers[0].PubE,
	)
	s.Require().NoError(err)

	assignedMembers := tsstypes.AssignedMembers(sa.AssignedMembers)
	signature, err := tsstestutils.GenerateSignature(
		signing,
		assignedMembers,
		tss.MemberID(1),
		de,
		s.tssGroupCtx.OwnPrivKeys[0],
	)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitSignature(
		&txOwner,
		signingID,
		tss.MemberID(1),
		signature,
	)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "tssLane", 2500000, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDKGRound1ExceedGas tests that MsgSubmitDKGRound1 transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDKGRound1ExceedGas() {
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDKGRound1(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.Round1Infos[0],
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDKGRound2ExceedGas tests that MsgSubmitDKGRound2 transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDKGRound2ExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDKGRound2(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.EncryptedSecretShares[0],
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgConfirmExceedGas tests that MsgConfirm transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgConfirmExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgConfirm(
		&txOwner,
		s.tssGroupCtx.GroupID,
		tss.MemberID(1),
		s.tssGroupCtx.OwnPubKeySigs[0],
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgComplainExceedGas tests that MsgComplain transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgComplainExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	// Note: complainantID and respondentID are memberID which is 1-indexed.
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg, err := genMsgComplain(
		&txOwner,
		s.tssGroupCtx.GroupID,
		s.tssGroupCtx.Round1Infos,
		1,
		2,
	)
	s.Require().NoError(err)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitDEExceedGas tests that MsgSubmitDE transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitDEExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	de := tsstestutils.GenerateDE(s.tssGroupCtx.Secrets[0])
	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitDE(
		&txOwner,
		[]tsstestutils.DEWithPrivateNonce{de},
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// TestTSSLaneWrappedMsgSubmitSignatureExceedGas tests that MsgSubmitSignature transactions with wrapped messages
// are rejected due to exceeding gas limit
func (s *AppTestSuite) TestTSSLaneWrappedMsgSubmitSignatureExceedGas() {
	err := s.tssGroupCtx.SubmitRound1(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound2(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.SubmitRound3(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	err = s.tssGroupCtx.FillDEs(s.ctx, s.app.TSSKeeper)
	s.Require().NoError(err)

	originator := tsstypes.NewDirectOriginator(
		"targetChain",
		"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		"test",
	)

	signingID, err := s.app.TSSKeeper.RequestSigning(
		s.ctx,
		s.tssGroupCtx.GroupID,
		&originator,
		tsstypes.NewTextSignatureOrder([]byte("msg")),
	)
	s.Require().NoError(err)

	signing, err := s.app.TSSKeeper.GetSigning(s.ctx, signingID)
	s.Require().NoError(err)

	sa, err := s.app.TSSKeeper.GetSigningAttempt(s.ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)

	de, err := s.tssGroupCtx.PopDE(
		0,
		sa.AssignedMembers[0].PubD,
		sa.AssignedMembers[0].PubE,
	)
	s.Require().NoError(err)

	assignedMembers := tsstypes.AssignedMembers(sa.AssignedMembers)
	signature, err := tsstestutils.GenerateSignature(
		signing,
		assignedMembers,
		tss.MemberID(1),
		de,
		s.tssGroupCtx.OwnPrivKeys[0],
	)
	s.Require().NoError(err)

	txOwner := s.tssAccountsWithNumSeq[0]
	sender := s.tssGranteeAccWithNumSeq
	msg := genMsgSubmitSignature(
		&txOwner,
		signingID,
		tss.MemberID(1),
		signature,
	)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "tssLane", 2500001, sdk.Coins{})
}

// -----------------------------------------------
// OracleReportLane tests
// -----------------------------------------------

// TestOracleReportLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestOracleReportLaneZeroGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgReportData(&sender)

	s.checkLaneWithZeroGas(msg, sender)
}

// TestOracleReportLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestOracleReportLaneExactGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgReportData(&sender)

	s.checkLaneWithExactGas(msg, sender, "oracleReportLane", 2500000, sdk.Coins{})
}

// TestOracleReportLaneExceedGas tests that transactions with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleReportLaneExceedGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgReportData(&sender)

	s.checkLaneWithExceedGas(msg, sender, "oracleReportLane", 2500001, sdk.Coins{})
}

// TestOracleReportLaneWrappedMsgExactGas tests that transactions with wrapped message with exact gas limit are accepted
func (s *AppTestSuite) TestOracleReportLaneWrappedMsgExactGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.reporterAccWithNumSeq
	msg := genMsgReportData(&txOwner)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "oracleReportLane", 2500000, sdk.Coins{})
}

// TestOracleReportLaneWrappedMsgExceedGas tests that transactions with wrapped message with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleReportLaneWrappedMsgExceedGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.reporterAccWithNumSeq
	msg := genMsgReportData(&txOwner)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "oracleReportLane", 2500001, sdk.Coins{})
}

// -----------------------------------------------
// OracleRequestLane tests
// -----------------------------------------------

// TestOracleRequestLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestOracleRequestLaneZeroGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgRequestData(&sender)

	s.checkLaneWithZeroGas(msg, sender)
}

// TestOracleRequestLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestOracleRequestLaneExactGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgRequestData(&sender)

	s.checkLaneWithExactGas(msg, sender, "oracleRequestLane", 5000000, sdk.Coins{sdk.NewInt64Coin("uband", 12500)})
}

// TestOracleRequestLaneExceedGas tests that transactions with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleRequestLaneExceedGas() {
	sender := s.valAccWithNumSeq
	msg := genMsgRequestData(&sender)

	s.checkLaneWithExceedGas(msg, sender, "oracleRequestLane", 5000001, sdk.Coins{sdk.NewInt64Coin("uband", 12501)})
}

// TestOracleRequestLaneWrappedMsgExactGas tests that transactions with wrapped message with exact gas limit are accepted
func (s *AppTestSuite) TestOracleRequestLaneWrappedMsgExactGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.reporterAccWithNumSeq
	msg := genMsgRequestData(&txOwner)

	s.checkLaneWrappedMsgWithExactGas(msg, sender, "oracleRequestLane", 5000000, sdk.Coins{sdk.NewInt64Coin("uband", 12500)})
}

// TestOracleRequestLaneWrappedMsgExceedGas tests that transactions with wrapped message with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleRequestLaneWrappedMsgExceedGas() {
	txOwner := s.valAccWithNumSeq
	sender := s.reporterAccWithNumSeq
	msg := genMsgRequestData(&txOwner)

	s.checkLaneWrappedMsgWithExceedGas(msg, sender, "oracleRequestLane", 5000001, sdk.Coins{sdk.NewInt64Coin("uband", 12501)})
}

// -----------------------------------------------
// Multiple lanes tests
// -----------------------------------------------

// TestRequestLaneBlockedByReportLane tests that request lane is blocked by report lane
func (s *AppTestSuite) TestRequestLaneBlockedByReportLane() {
	require := s.Require()

	// Generate report data transaction
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgReportData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		2500000,
		4,
	)

	// Check that the report data transactions are accepted
	for _, txBytes := range reportTxBytes {
		checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
		res, err := s.app.CheckTx(checkTxReq)
		require.NoError(err)
		require.NotNil(res)
		require.Equal(res.Code, uint32(0))
	}

	require.Equal(s.app.Mempool().CountTx(), 4)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 4)

	// Generate request data transaction
	requestTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgRequestData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12500),
		},
		5000000,
		1,
	)

	// Check that the request data transactions are accepted
	for _, txBytes := range requestTxBytes {
		checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
		res, err := s.app.CheckTx(checkTxReq)
		require.NoError(err)
		require.NotNil(res)
		require.Equal(res.Code, uint32(0))
	}

	require.Equal(s.app.Mempool().CountTx(), 5)
	require.Equal(mempool.GetLane("oracleRequestLane").CountTx(), 1)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 4)
	require.Equal(resp.Txs, reportTxBytes)
}

// TestAllLaneFilled tests fill all lanes by their limit
func (s *AppTestSuite) TestAllLaneFilled() {
	require := s.Require()

	// Generate bank send transaction
	bankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 12500)},
		5000000,
		1,
	)

	s.checkTxAcceptance(bankSendTxBytes, "defaultLane")

	// Generate request data transaction
	requestTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgRequestData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12500),
		},
		2500000,
		2,
	)

	// Check that the request data transactions are accepted
	s.checkTxAcceptance(requestTxBytes, "oracleRequestLane")

	// Generate report data transaction
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgReportData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		2500000,
		4,
	)

	// Check that the report data transactions are accepted
	s.checkTxAcceptance(reportTxBytes, "oracleReportLane")

	// Generate tss transaction
	tssTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSubmitDKGRound1(
				&s.tssAccountsWithNumSeq[0],
				s.tssGroupCtx.GroupID,
				tss.MemberID(1),
				s.tssGroupCtx.Round1Infos[0],
			),
		},
		&s.tssAccountsWithNumSeq[0],
		sdk.Coins{},
		2500000,
		4,
	)

	s.checkTxAcceptance(tssTxBytes, "tssLane")

	// Generate feeds transaction
	feedsTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		1000000,
		25,
	)

	s.checkTxAcceptance(feedsTxBytes, "feedsLane")

	require.Equal(s.app.Mempool().CountTx(), 36)
	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("defaultLane").CountTx(), 1)
	require.Equal(mempool.GetLane("oracleRequestLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 4)
	require.Equal(mempool.GetLane("tssLane").CountTx(), 4)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 25)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}

	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 34)
	require.Equal(resp.Txs, append(append(append(feedsTxBytes, tssTxBytes...), reportTxBytes...), bankSendTxBytes...)) // every lane except oracleRequestLane
}

// TestAllLaneFilledExceptOracleReportLane tests fill all lanes by their limit except oracle report lane
func (s *AppTestSuite) TestAllLaneFilledExceptOracleReportLane() {
	require := s.Require()

	// Generate bank send transaction
	bankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 6250)},
		2500000,
		2,
	)

	s.checkTxAcceptance(bankSendTxBytes, "defaultLane")

	// Generate request data transaction
	requestTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgRequestData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{
			sdk.NewInt64Coin("uband", 6250),
		},
		2500000,
		2,
	)

	// Check that the request data transactions are accepted
	s.checkTxAcceptance(requestTxBytes, "oracleRequestLane")

	// Generate report data transaction
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgReportData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		2500000,
		3,
	)

	// Check that the report data transactions are accepted
	s.checkTxAcceptance(reportTxBytes, "oracleReportLane")

	// Generate tss transaction
	tssTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSubmitDKGRound1(
				&s.tssAccountsWithNumSeq[0],
				s.tssGroupCtx.GroupID,
				tss.MemberID(1),
				s.tssGroupCtx.Round1Infos[0],
			),
		},
		&s.tssAccountsWithNumSeq[0],
		sdk.Coins{},
		2500000,
		4,
	)

	s.checkTxAcceptance(tssTxBytes, "tssLane")

	// Generate feeds transaction
	feedsTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		1000000,
		25,
	)

	s.checkTxAcceptance(feedsTxBytes, "feedsLane")

	require.Equal(s.app.Mempool().CountTx(), 36)
	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("defaultLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleRequestLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 3)
	require.Equal(mempool.GetLane("tssLane").CountTx(), 4)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 25)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}

	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 35)
	// Since the OracleReportLane is not filled, the OracleRequestLane is included
	require.Equal(resp.Txs, append(append(append(append(feedsTxBytes, tssTxBytes...), reportTxBytes...), requestTxBytes...), bankSendTxBytes[0])) // only the first bank send transaction is included
}

// TestAllLaneFilledExceed tests fill all lanes exceed their limit
func (s *AppTestSuite) TestAllLaneFilledExceed() {
	require := s.Require()

	// Generate bank send transaction
	bankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 12500)},
		4999999,
		2,
	)

	s.checkTxAcceptance(bankSendTxBytes, "defaultLane")

	// Generate request data transaction
	requestTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgRequestData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12500),
		},
		4999999,
		2,
	)

	// Check that the request data transactions are accepted
	s.checkTxAcceptance(requestTxBytes, "oracleRequestLane")

	// Generate report data transaction
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgReportData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		2499999,
		5,
	)

	// Check that the report data transactions are accepted
	s.checkTxAcceptance(reportTxBytes, "oracleReportLane")

	// Generate tss transaction
	tssTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSubmitDKGRound1(
				&s.tssAccountsWithNumSeq[0],
				s.tssGroupCtx.GroupID,
				tss.MemberID(1),
				s.tssGroupCtx.Round1Infos[0],
			),
		},
		&s.tssAccountsWithNumSeq[0],
		sdk.Coins{},
		2499999,
		5,
	)

	s.checkTxAcceptance(tssTxBytes, "tssLane")

	// Generate feeds transaction
	feedsTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		999999,
		26,
	)

	s.checkTxAcceptance(feedsTxBytes, "feedsLane")

	require.Equal(s.app.Mempool().CountTx(), 40)
	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("defaultLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleRequestLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 5)
	require.Equal(mempool.GetLane("tssLane").CountTx(), 5)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 26)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}

	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 35)
	require.Equal(resp.Txs, append(append(feedsTxBytes, tssTxBytes...), reportTxBytes[0:4]...))
}

// TestFillRemainingProposal tests fill the remaining proposal
func (s *AppTestSuite) TestFillRemainingProposal() {
	require := s.Require()

	// Generate bank send transaction
	bankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 12500)},
		5000000,
		3,
	)

	s.checkTxAcceptance(bankSendTxBytes, "defaultLane")

	// Generate feeds transaction
	feedsTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		1000000,
		40,
	)

	s.checkTxAcceptance(feedsTxBytes, "feedsLane")

	require.Equal(s.app.Mempool().CountTx(), 43)
	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("defaultLane").CountTx(), 3)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 40)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}

	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 42)

	var expectedTxBytes [][]byte
	expectedTxBytes = append(append(append(append(expectedTxBytes, feedsTxBytes[:25]...), bankSendTxBytes[0]), feedsTxBytes[25:]...), bankSendTxBytes[1])

	require.Equal(resp.Txs, expectedTxBytes)
}

// TestLargeTxSizeBlocksSubsequentTx tests that a large transaction size blocks subsequent transactions
func (s *AppTestSuite) TestLargeTxSizeBlocksSubsequentTx() {
	require := s.Require()

	// Generate large bank send transaction
	largeBankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 12500)},
		5000000,
		1,
	)

	s.checkTxAcceptance(largeBankSendTxBytes, "defaultLane")

	// Generate small bank send transaction
	smallBankSendTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSend(&s.valAccWithNumSeq, &s.reporterAccWithNumSeq, sdk.Coins{sdk.NewInt64Coin("uband", 1)}),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{sdk.NewInt64Coin("uband", 250)},
		100000,
		1,
	)

	s.checkTxAcceptance(smallBankSendTxBytes, "defaultLane")

	// Generate report data transaction
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgReportData(&s.valAccWithNumSeq),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		2500000,
		4,
	)

	// Check that the report data transactions are accepted
	s.checkTxAcceptance(reportTxBytes, "oracleReportLane")

	// Generate tss transaction
	tssTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			genMsgSubmitDKGRound1(
				&s.tssAccountsWithNumSeq[0],
				s.tssGroupCtx.GroupID,
				tss.MemberID(1),
				s.tssGroupCtx.Round1Infos[0],
			),
		},
		&s.tssAccountsWithNumSeq[0],
		sdk.Coins{},
		2500000,
		4,
	)

	s.checkTxAcceptance(tssTxBytes, "tssLane")

	// Generate feeds transaction
	feedsTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		&s.valAccWithNumSeq,
		sdk.Coins{},
		999999,
		28,
	)

	s.checkTxAcceptance(feedsTxBytes, "feedsLane")

	require.Equal(s.app.Mempool().CountTx(), 38)
	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("defaultLane").CountTx(), 2)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 4)
	require.Equal(mempool.GetLane("tssLane").CountTx(), 4)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 28)

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}

	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 36)

	var expectedTxBytes [][]byte
	// the bank send txs are blocked by the large bank send tx
	expectedTxBytes = append(append(append(append(expectedTxBytes, feedsTxBytes[:26]...), tssTxBytes...), reportTxBytes...), feedsTxBytes[26:]...)
	require.Equal(resp.Txs, expectedTxBytes)
}

// -----------------------------------------------
// Helper functions
// -----------------------------------------------

func (s *AppTestSuite) checkLaneWithZeroGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		msgs,
		sdk.Coins{},
		0,
		bandtesting.ChainID,
		[]uint64{sender.Num},
		[]uint64{sender.Seq},
		sender.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(11), res.Code)
	require.Equal(0, s.app.Mempool().CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(0, len(resp.Txs))
}

func (s *AppTestSuite) checkLaneWithExactGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq, laneName string, gas uint64, feeAmt sdk.Coins) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		msgs,
		feeAmt,
		gas,
		bandtesting.ChainID,
		[]uint64{sender.Num},
		[]uint64{sender.Seq},
		sender.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	fmt.Println("res", res)
	require.Equal(uint32(0), res.Code)
	require.Equal(1, s.app.Mempool().CountTx())

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(1, mempool.GetLane(laneName).CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(len(resp.Txs), 1)
	require.Equal(resp.Txs[0], txBytes)
}

func (s *AppTestSuite) checkLaneWithExceedGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq, laneName string, gas uint64, feeAmt sdk.Coins) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		msgs,
		feeAmt,
		gas,
		bandtesting.ChainID,
		[]uint64{sender.Num},
		[]uint64{sender.Seq},
		sender.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(1), res.Code)
	require.Equal(0, s.app.Mempool().CountTx())

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(0, mempool.GetLane(laneName).CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(0, len(resp.Txs))
}

func (s *AppTestSuite) checkLaneWrappedMsgWithExactGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq, laneName string, gas uint64, feeAmt sdk.Coins) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(sender.Address, msgs)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		feeAmt,
		gas,
		bandtesting.ChainID,
		[]uint64{sender.Num},
		[]uint64{sender.Seq},
		sender.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(0), res.Code)
	require.Equal(1, s.app.Mempool().CountTx())

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(1, mempool.GetLane(laneName).CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(1, len(resp.Txs))
	require.Equal(txBytes, resp.Txs[0])
}

func (s *AppTestSuite) checkLaneWrappedMsgWithExceedGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq, laneName string, gas uint64, feeAmt sdk.Coins) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(sender.Address, msgs)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		feeAmt,
		gas,
		bandtesting.ChainID,
		[]uint64{sender.Num},
		[]uint64{sender.Seq},
		sender.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(uint32(1), res.Code)
	require.Equal(0, s.app.Mempool().CountTx())

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(0, mempool.GetLane(laneName).CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      1000000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(0, len(resp.Txs))
}

// checkTxAcceptance checks that the transaction is accepted
func (s *AppTestSuite) checkTxAcceptance(txBytes [][]byte, laneName string) {
	require := s.Require()

	mempoolCountBefore := s.app.Mempool().CountTx()
	mempool := s.app.Mempool().(*mempool.Mempool)
	laneCountBefore := mempool.GetLane(laneName).CountTx()

	for _, txBytes := range txBytes {
		checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
		res, err := s.app.CheckTx(checkTxReq)
		require.NoError(err)
		require.NotNil(res)
		require.Equal(res.Code, uint32(0))
	}

	require.Equal(mempoolCountBefore+len(txBytes), mempool.CountTx())
	require.Equal(laneCountBefore+len(txBytes), mempool.GetLane(laneName).CountTx())
}

// genMsgSend creates a message for sending coins
func genMsgSend(sender *bandtesting.AccountWithNumSeq, recipient *bandtesting.AccountWithNumSeq, amount sdk.Coins) sdk.Msg {
	return banktypes.NewMsgSend(sender.Address, recipient.Address, amount)
}

// genMsgRequestData creates a message for requesting data
func genMsgRequestData(sender *bandtesting.AccountWithNumSeq) sdk.Msg {
	return oracletypes.NewMsgRequestData(
		oracletypes.OracleScriptID(1),
		[]byte("calldata"),
		1,
		1,
		"app_test",
		sdk.NewCoins(sdk.NewInt64Coin("uband", 9000000)),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		sender.Address,
		0,
	)
}

// genMsgReportData creates a message for reporting data
func genMsgReportData(sender *bandtesting.AccountWithNumSeq) sdk.Msg {
	return oracletypes.NewMsgReportData(
		oracletypes.RequestID(1), []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("answer1")),
			oracletypes.NewRawReport(2, 0, []byte("answer2")),
			oracletypes.NewRawReport(3, 0, []byte("answer3")),
		},
		sender.ValAddress,
	)
}

func genMsgSubmitDKGRound1(
	sender *bandtesting.AccountWithNumSeq,
	groupId tss.GroupID,
	memberID tss.MemberID,
	round1Info tss.Round1Info,
) sdk.Msg {
	return tsstypes.NewMsgSubmitDKGRound1(
		groupId,
		tsstypes.Round1Info{
			MemberID:           memberID,
			CoefficientCommits: round1Info.CoefficientCommits,
			OneTimePubKey:      round1Info.OneTimePubKey,
			A0Signature:        round1Info.A0Signature,
			OneTimeSignature:   round1Info.OneTimeSignature,
		},
		sender.Address.String(),
	)
}

func genMsgSubmitDKGRound2(
	sender *bandtesting.AccountWithNumSeq,
	groupID tss.GroupID,
	memberID tss.MemberID,
	encryptedSecretShares tss.EncSecretShares,
) sdk.Msg {
	return tsstypes.NewMsgSubmitDKGRound2(
		groupID,
		tsstypes.Round2Info{
			MemberID:              memberID,
			EncryptedSecretShares: encryptedSecretShares,
		},
		sender.Address.String(),
	)
}

func genMsgConfirm(
	sender *bandtesting.AccountWithNumSeq,
	groupID tss.GroupID,
	memberID tss.MemberID,
	ownPubKey tss.Signature,
) sdk.Msg {
	return tsstypes.NewMsgConfirm(
		groupID,
		memberID,
		ownPubKey,
		sender.Address.String(),
	)
}

func genMsgComplain(
	sender *bandtesting.AccountWithNumSeq,
	groupID tss.GroupID,
	round1Infos []tss.Round1Info,
	complainantID int,
	respondentID int,
) (sdk.Msg, error) {
	signature, keySym, err := tss.SignComplaint(
		round1Infos[complainantID].OneTimePubKey,
		round1Infos[respondentID].OneTimePubKey,
		round1Infos[complainantID].OneTimePrivKey,
	)
	if err != nil {
		return nil, err
	}

	return tsstypes.NewMsgComplain(
		groupID,
		[]tsstypes.Complaint{
			{
				Complainant: tss.MemberID(complainantID),
				Respondent:  tss.MemberID(respondentID),
				KeySym:      keySym,
				Signature:   signature,
			},
		},
		sender.Address.String(),
	), nil
}

func genMsgSubmitDE(
	sender *bandtesting.AccountWithNumSeq,
	deWithPrivateNonces []tsstestutils.DEWithPrivateNonce,
) sdk.Msg {
	des := make([]tsstypes.DE, len(deWithPrivateNonces))
	for i, de := range deWithPrivateNonces {
		des[i] = tsstypes.DE{
			PubD: de.PubDE.PubD,
			PubE: de.PubDE.PubE,
		}
	}

	return tsstypes.NewMsgSubmitDEs(des, sender.Address.String())
}

func genMsgSubmitSignature(
	sender *bandtesting.AccountWithNumSeq,
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
) sdk.Msg {
	return tsstypes.NewMsgSubmitSignature(signingID, memberID, signature, sender.Address.String())
}

func GenMsgSubmitSignalPrices(
	sender *bandtesting.AccountWithNumSeq,
	feeds []feedstypes.Feed,
	timestamp int64,
) sdk.Msg {
	prices := make([]feedstypes.SignalPrice, 0, len(feeds))
	for _, feed := range feeds {
		prices = append(prices, feedstypes.SignalPrice{
			Status:   feedstypes.SIGNAL_PRICE_STATUS_AVAILABLE,
			SignalID: feed.SignalID,
			Price:    60000,
		})
	}

	return feedstypes.NewMsgSubmitSignalPrices(sender.ValAddress.String(), timestamp, prices)
}

// GenFeeds a number of feeds
func GenFeeds(num int) (feeds []feedstypes.Feed) {
	for i := range num {
		feeds = append(feeds, feedstypes.Feed{
			SignalID: fmt.Sprintf("signal.%d", i),
			Power:    int64(60_000_000_000),
			Interval: 60,
		})
	}

	return
}
