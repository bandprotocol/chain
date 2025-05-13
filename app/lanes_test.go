package band_test

import (
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
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

type AppTestSuite struct {
	suite.Suite

	app      *band.BandApp
	ctx      sdk.Context
	txConfig client.TxConfig

	valAccWithNumSeq      bandtesting.AccountWithNumSeq
	feederAccWithNumSeq   bandtesting.AccountWithNumSeq
	reporterAccWithNumSeq bandtesting.AccountWithNumSeq
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

// TestTSSLaneZeroGas tests that transactions with zero gas are rejected
// func (s *AppTestSuite) TestTSSLaneZeroGas() {
// }

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
