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
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

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

// -----------------------------------------------
// FeedsLane tests
// -----------------------------------------------

// TestFeedsLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestFeedsLaneZeroGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		sdk.Coins{},
		0,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(11))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestFeedsLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestFeedsLaneExactGas() {
	require := s.Require()

	msg := GenMsgSubmitSignalPrices(
		&s.valAccWithNumSeq,
		s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
		s.ctx.BlockTime().Unix(),
	)

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			msg,
		},
		sdk.Coins{},
		1000000,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 1)

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

// TestFeedsLaneExceedGas tests that transactions with gas exceeding limit are rejected
func (s *AppTestSuite) TestFeedsLaneExceedGas() {
	require := s.Require()

	// Tx with gas greater than the tx gas limit of lane
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
		sdk.Coins{},
		1000001,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestFeedsLaneWrappedMsgExactGas tests that transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestFeedsLaneWrappedMsgExactGas() {
	require := s.Require()

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(
		s.feederAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			&msgExec,
		},
		sdk.Coins{},
		1000000,
		bandtesting.ChainID,
		[]uint64{s.feederAccWithNumSeq.Num},
		[]uint64{s.feederAccWithNumSeq.Seq},
		s.feederAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 1)

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

// TestFeedsLaneWrappedMsgExceedGas tests that transactions with wrapped messages are handled correctly
func (s *AppTestSuite) TestFeedsLaneWrappedMsgExceedGas() {
	require := s.Require()

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(
		s.feederAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgSubmitSignalPrices(
				&s.valAccWithNumSeq,
				s.app.FeedsKeeper.GetCurrentFeeds(s.ctx).Feeds,
				s.ctx.BlockTime().Unix(),
			),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			&msgExec,
		},
		sdk.Coins{},
		1000001,
		bandtesting.ChainID,
		[]uint64{s.feederAccWithNumSeq.Num},
		[]uint64{s.feederAccWithNumSeq.Seq},
		s.feederAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("feedsLane").CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
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

	s.checkTSSLaneZeroGas(msg, sender)
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

	s.checkTSSLaneZeroGas(msg, sender)
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

	s.checkTSSLaneZeroGas(msg, sender)
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

	s.checkTSSLaneZeroGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneZeroGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneZeroGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExactGas(msg, sender)
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

	s.checkTSSLaneExactGas(msg, sender)
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

	s.checkTSSLaneExactGas(msg, sender)
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

	s.checkTSSLaneExactGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExactGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExactGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExceedGas(msg, sender)
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

	s.checkTSSLaneExceedGas(msg, sender)
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

	s.checkTSSLaneExceedGas(msg, sender)
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

	s.checkTSSLaneExceedGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExceedGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneExceedGas(msg, s.tssAccountsWithNumSeq[0])
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExactGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
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

	s.checkTSSLaneWrappedMsgExceedGas(msg, sender)
}

func (s *AppTestSuite) checkTSSLaneZeroGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
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

func (s *AppTestSuite) checkTSSLaneExactGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		msgs,
		sdk.Coins{},
		2500000,
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
	require.Equal(1, mempool.GetLane("tssLane").CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      2500000,
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

func (s *AppTestSuite) checkTSSLaneExceedGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		msgs,
		sdk.Coins{},
		2500001,
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
	require.Equal(0, mempool.GetLane("tssLane").CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      2500001,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(0, len(resp.Txs))
}

func (s *AppTestSuite) checkTSSLaneWrappedMsgExactGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(sender.Address, msgs)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{},
		2500000,
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
	require.Equal(1, mempool.GetLane("tssLane").CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      2500000,
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

func (s *AppTestSuite) checkTSSLaneWrappedMsgExceedGas(msg sdk.Msg, sender bandtesting.AccountWithNumSeq) {
	require := s.Require()
	msgs := []sdk.Msg{msg}

	// Tx with msg wrapped in msg Exec
	msgExec := authz.NewMsgExec(sender.Address, msgs)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{},
		2500001,
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
	require.Equal(0, mempool.GetLane("tssLane").CountTx())

	// Prepare proposal
	prepareReq := &abci.RequestPrepareProposal{
		MaxTxBytes:      2500000,
		Height:          s.app.LastBlockHeight() + 1,
		Time:            s.ctx.BlockTime(),
		ProposerAddress: bandtesting.Validators[0].Address.Bytes(),
	}
	resp, err := s.app.PrepareProposal(prepareReq)
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(0, len(resp.Txs))
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

// -----------------------------------------------
// OracleReportLane tests
// -----------------------------------------------

// TestOracleReportLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestOracleReportLaneZeroGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
		},
		sdk.Coins{},
		0,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(11))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestOracleReportLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestOracleReportLaneExactGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
		},
		sdk.Coins{},
		2500000,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 1)

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

// TestOracleReportLaneExceedGas tests that transactions with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleReportLaneExceedGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
		},
		sdk.Coins{},
		2500001,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestOracleReportLaneWrappedMsgExactGas tests that transactions with wrapped message with exact gas limit are accepted
func (s *AppTestSuite) TestOracleReportLaneWrappedMsgExactGas() {
	require := s.Require()

	msgExec := authz.NewMsgExec(
		s.reporterAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{},
		2500000,
		bandtesting.ChainID,
		[]uint64{s.reporterAccWithNumSeq.Num},
		[]uint64{s.reporterAccWithNumSeq.Seq},
		s.reporterAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

	mempool := s.app.Mempool().(*mempool.Mempool)
	require.Equal(mempool.GetLane("oracleReportLane").CountTx(), 1)

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

// TestOracleReportLaneWrappedMsgExceedGas tests that transactions with wrapped message with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleReportLaneWrappedMsgExceedGas() {
	require := s.Require()

	msgExec := authz.NewMsgExec(
		s.reporterAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{},
		2500001,
		bandtesting.ChainID,
		[]uint64{s.reporterAccWithNumSeq.Num},
		[]uint64{s.reporterAccWithNumSeq.Seq},
		s.reporterAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// GenMsgReportData creates a message for reporting data
func GenMsgReportData(sender *bandtesting.AccountWithNumSeq) sdk.Msg {
	return oracletypes.NewMsgReportData(
		oracletypes.RequestID(1), []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("answer1")),
			oracletypes.NewRawReport(2, 0, []byte("answer2")),
			oracletypes.NewRawReport(3, 0, []byte("answer3")),
		},
		sender.ValAddress,
	)
}

// -----------------------------------------------
// OracleRequestLane tests
// -----------------------------------------------

// TestOracleRequestLaneZeroGas tests that transactions with zero gas are rejected
func (s *AppTestSuite) TestOracleRequestLaneZeroGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
		},
		sdk.Coins{},
		0,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(11))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestOracleRequestLaneExactGas tests that transactions with exact gas limit are accepted
func (s *AppTestSuite) TestOracleRequestLaneExactGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
		},
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12500),
		},
		5000000,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

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

// TestOracleRequestLaneExceedGas tests that transactions with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleRequestLaneExceedGas() {
	require := s.Require()

	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
		},
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12501),
		},
		5000001,
		bandtesting.ChainID,
		[]uint64{s.valAccWithNumSeq.Num},
		[]uint64{s.valAccWithNumSeq.Seq},
		s.valAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// TestOracleRequestLaneWrappedMsgExactGas tests that transactions with wrapped message with exact gas limit are accepted
func (s *AppTestSuite) TestOracleRequestLaneWrappedMsgExactGas() {
	require := s.Require()

	msgExec := authz.NewMsgExec(
		s.reporterAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12500),
		},
		5000000,
		bandtesting.ChainID,
		[]uint64{s.reporterAccWithNumSeq.Num},
		[]uint64{s.reporterAccWithNumSeq.Seq},
		s.reporterAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(0))
	require.Equal(s.app.Mempool().CountTx(), 1)

	mempool := s.app.Mempool().(*mempool.Mempool)
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
	require.Equal(len(resp.Txs), 1)
	require.Equal(resp.Txs[0], txBytes)
}

// TestOracleRequestLaneWrappedMsgExceedGas tests that transactions with wrapped message with exceed gas limit are rejected
func (s *AppTestSuite) TestOracleRequestLaneWrappedMsgExceedGas() {
	require := s.Require()

	msgExec := authz.NewMsgExec(
		s.reporterAccWithNumSeq.Address,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
		},
	)
	tx, _ := bandtesting.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		s.txConfig,
		[]sdk.Msg{&msgExec},
		sdk.Coins{
			sdk.NewInt64Coin("uband", 12501),
		},
		5000001,
		bandtesting.ChainID,
		[]uint64{s.reporterAccWithNumSeq.Num},
		[]uint64{s.reporterAccWithNumSeq.Seq},
		s.reporterAccWithNumSeq.PrivKey,
	)

	txBytes, err := s.txConfig.TxEncoder()(tx)
	require.NoError(err)

	checkTxReq := &abci.RequestCheckTx{Tx: txBytes, Type: abci.CheckTxType_New}
	res, err := s.app.CheckTx(checkTxReq)
	require.NoError(err)
	require.NotNil(res)
	require.Equal(res.Code, uint32(1))
	require.Equal(s.app.Mempool().CountTx(), 0)

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
	require.Equal(len(resp.Txs), 0)
}

// GenMsgRequestData creates a message for requesting data
func GenMsgRequestData(sender *bandtesting.AccountWithNumSeq) sdk.Msg {
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

// -----------------------------------------------
// Multiple lanes tests
// -----------------------------------------------

// TestRequestLaneBlockedByReportLane tests that request lane is blocked by report lane
func (s *AppTestSuite) TestRequestLaneBlockedByReportLane() {
	require := s.Require()

	// Generate 4 report data transactions
	reportTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgReportData(&s.valAccWithNumSeq),
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

	// Generate 4 request data transactions
	requestTxBytes := bandtesting.GenSequenceOfTxs(
		s.txConfig.TxEncoder(),
		s.txConfig,
		[]sdk.Msg{
			GenMsgRequestData(&s.valAccWithNumSeq),
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
