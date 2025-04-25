package mempool

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// LaneTestSuite is a testify.Suite for unit-testing the Lane functionality.
type LaneTestSuite struct {
	suite.Suite

	encodingConfig EncodingConfig
	random         *rand.Rand
	accounts       []Account
	gasTokenDenom  string
	ctx            sdk.Context
}

func TestLaneTestSuite(t *testing.T) {
	suite.Run(t, new(LaneTestSuite))
}

func (s *LaneTestSuite) SetupTest() {
	s.encodingConfig = CreateTestEncodingConfig()
	s.random = rand.New(rand.NewSource(1))
	s.accounts = RandomAccounts(s.random, 3)
	s.gasTokenDenom = "uband"

	testCtx := testutil.DefaultContextWithDB(
		s.T(),
		storetypes.NewKVStoreKey("test"),
		storetypes.NewTransientStoreKey("transient_test"),
	)
	s.ctx = testCtx.Ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithBlockHeight(1)
	s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 1000000000000,
			MaxGas:   100,
		},
	})
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func (s *LaneTestSuite) TestLaneInsertAndCount() {
	// Create a Lane that matches all txs (Match always returns true),
	// just to test Insert/Count.
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// Create and insert two transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 10)
	tx2 := s.createSimpleTx(s.accounts[1], 0, 10)

	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))

	// Ensure lane sees 2 transactions
	s.Require().Equal(2, lane.CountTx())

	// Create over gas limit tx
	tx3 := s.createSimpleTx(s.accounts[2], 0, 100)
	s.Require().Error(lane.Insert(s.ctx, tx3))

	// Ensure lane does not insert tx3
	s.Require().Equal(2, lane.CountTx())

	// set bytes limit to 1000
	s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 500,
			MaxGas:   100,
		},
	})

	// Create over bytes limit tx
	tx4 := s.createSimpleTx(s.accounts[2], 0, 0) // 217 bytes
	s.Require().Error(lane.Insert(s.ctx, tx4))

	// Ensure lane does not insert tx4
	s.Require().Equal(2, lane.CountTx())
}

func (s *LaneTestSuite) TestLaneRemove() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	tx := s.createSimpleTx(s.accounts[0], 0, 10)
	s.Require().NoError(lane.Insert(s.ctx, tx))
	s.Require().Equal(1, lane.CountTx())

	// Remove it
	err := lane.Remove(tx)
	s.Require().NoError(err)
	s.Require().Equal(0, lane.CountTx())
}

func (s *LaneTestSuite) TestLaneFillProposalWithGasLimit() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// Insert 3 transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 20)
	tx2 := s.createSimpleTx(s.accounts[1], 1, 20)
	tx3 := s.createSimpleTx(s.accounts[2], 2, 20)
	tx4 := s.createSimpleTx(s.accounts[2], 3, 20)
	tx5 := s.createSimpleTx(s.accounts[2], 4, 15)
	tx6 := s.createSimpleTx(s.accounts[2], 5, 10)

	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))
	s.Require().NoError(lane.Insert(s.ctx, tx3))
	s.Require().NoError(lane.Insert(s.ctx, tx4))
	s.Require().NoError(lane.Insert(s.ctx, tx5))
	s.Require().NoError(lane.Insert(s.ctx, tx6))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator := lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 and tx2 to be included in the proposal.
	// Since the 20% of 1000 is 200, the gas should be over the limit, so tx3 is yet to be considered.
	s.Require().Equal(uint64(40), blockUsed.Gas(), "20 gas from tx1 and 20 gas from tx2")
	s.Require().NotNil(iterator)

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs := s.getTxBytes(tx1, tx2)
	s.Require().Equal(2, len(proposal.txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	// Calculate the remaining block space
	remainderLimit := proposal.maxBlockSpace.Sub(proposal.totalBlockSpace)

	// Call FillProposalBy with the remainder limit and iterator from the previous call.
	blockUsed = lane.FillProposalByIterator(&proposal, iterator, remainderLimit)

	// We expect tx1, tx2, tx3, tx4, tx5 to be included in the proposal.
	s.Require().Equal(uint64(55), blockUsed.Gas(), "20 gas from tx3 and 20 gas from tx4 + 15 gas from tx5")

	// The proposal should contain 5 transactions in Txs().
	expectedIncludedTxs = s.getTxBytes(tx1, tx2, tx3, tx4, tx5)
	s.Require().Equal(5, len(proposal.txs), "five txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)
}

func (s *LaneTestSuite) TestLaneFillProposalWithBytesLimit() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	// Insert 3 transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 0) // 217 bytes
	tx2 := s.createSimpleTx(s.accounts[1], 1, 0) // 219 bytes
	tx3 := s.createSimpleTx(s.accounts[2], 2, 0) // 219 bytes
	tx4 := s.createSimpleTx(s.accounts[2], 3, 0) // 219 bytes
	tx5 := s.createSimpleTx(s.accounts[2], 4, 0) // 219 bytes

	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))
	s.Require().NoError(lane.Insert(s.ctx, tx3))
	s.Require().NoError(lane.Insert(s.ctx, tx4))
	s.Require().NoError(lane.Insert(s.ctx, tx5))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000,
		1000000000000,
	)

	// FillProposal
	blockUsed, iterator := lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 and tx2 to be included in the proposal.
	// Since the 30% of 1000 is 300, the bytes should be over the limit, so tx3 is yet to be considered.
	s.Require().Equal(uint64(436), blockUsed.TxBytes())

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs := s.getTxBytes(tx1, tx2)
	s.Require().Equal(2, len(proposal.txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	// Calculate the remaining block space
	remainderLimit := proposal.maxBlockSpace.Sub(proposal.totalBlockSpace)

	// Call FillProposalBy with the remainder limit and iterator from the previous call.
	blockUsed = lane.FillProposalByIterator(&proposal, iterator, remainderLimit)

	// We expect tx1, tx2, tx3, tx4 to be included in the proposal.
	s.Require().Equal(uint64(438), blockUsed.TxBytes())

	// The proposal should contain 4 transactions in Txs().
	expectedIncludedTxs = s.getTxBytes(tx1, tx2, tx3, tx4)
	s.Require().Equal(4, len(proposal.txs), "four txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)
}

type callbackAfterFillProposalMock struct {
	isLaneLimitExceeded bool
}

func (f *callbackAfterFillProposalMock) callbackAfterFillProposal(isLaneLimitExceeded bool) {
	f.isLaneLimitExceeded = isLaneLimitExceeded
}

func (s *LaneTestSuite) TestLaneCallbackAfterFillProposal() {
	callbackAfterFillProposalMock := &callbackAfterFillProposalMock{}
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		callbackAfterFillProposalMock.callbackAfterFillProposal,
	)

	// Insert a transaction
	tx1 := s.createSimpleTx(s.accounts[0], 0, 20)

	s.Require().NoError(lane.Insert(s.ctx, tx1))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator := lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 to be included in the proposal.
	s.Require().Equal(uint64(20), blockUsed.Gas(), "20 gas from tx1")
	s.Require().Nil(iterator)

	// The proposal should contain 1 transaction in Txs().
	expectedIncludedTxs := s.getTxBytes(tx1)
	s.Require().Equal(1, len(proposal.txs), "one txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().False(callbackAfterFillProposalMock.isLaneLimitExceeded, "callbackAfterFillProposal should be called with false")

	// Insert 2 more transactions
	tx2 := s.createSimpleTx(s.accounts[1], 1, 20)
	tx3 := s.createSimpleTx(s.accounts[2], 2, 30)

	s.Require().NoError(lane.Insert(s.ctx, tx2))
	s.Require().NoError(lane.Insert(s.ctx, tx3))

	// Create a proposal with block-limits
	proposal = NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator = lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 and tx2 to be included in the proposal.
	// Then the gas should be over the limit, so tx3 is yet to be considered.
	s.Require().Equal(uint64(40), blockUsed.Gas(), "20 gas from tx1 and 20 gas from tx2")
	s.Require().NotNil(iterator)

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs = s.getTxBytes(tx1, tx2)
	s.Require().Equal(2, len(proposal.txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().True(callbackAfterFillProposalMock.isLaneLimitExceeded, "OoLaneLimitExceeded should be called with true")
}

func (s *LaneTestSuite) TestLaneExactlyFilled() {
	callbackAfterFillProposalMock := &callbackAfterFillProposalMock{}
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		callbackAfterFillProposalMock.callbackAfterFillProposal,
	)

	// Insert a transaction
	tx1 := s.createSimpleTx(s.accounts[0], 0, 20)

	s.Require().NoError(lane.Insert(s.ctx, tx1))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator := lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 to be included in the proposal.
	s.Require().Equal(uint64(20), blockUsed.Gas(), "20 gas from tx1")
	s.Require().Nil(iterator)

	// The proposal should contain 1 transaction in Txs().
	expectedIncludedTxs := s.getTxBytes(tx1)
	s.Require().Equal(1, len(proposal.txs), "one txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().False(callbackAfterFillProposalMock.isLaneLimitExceeded, "callbackAfterFillProposal should be called with false")

	// Insert 2 more transactions
	tx2 := s.createSimpleTx(s.accounts[1], 1, 10)
	tx3 := s.createSimpleTx(s.accounts[2], 2, 30)

	s.Require().NoError(lane.Insert(s.ctx, tx2))
	s.Require().NoError(lane.Insert(s.ctx, tx3))

	// Create a proposal with block-limits
	proposal = NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator = lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 and tx2 to be included in the proposal.
	// Then the gas should be over the limit, so tx3 is yet to be considered.
	s.Require().Equal(uint64(30), blockUsed.Gas(), "20 gas from tx1 and 10 gas from tx2")
	s.Require().NotNil(iterator)

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs = s.getTxBytes(tx1, tx2)
	s.Require().Equal(2, len(proposal.txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().True(callbackAfterFillProposalMock.isLaneLimitExceeded, "callbackAfterFillProposal should be called with true")
}

func (s *LaneTestSuite) TestLaneBlocked() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
		nil,
	)

	lane.SetBlocked(true)

	// Insert 3 transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 20)
	tx2 := s.createSimpleTx(s.accounts[1], 1, 20)

	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	blockUsed, iterator := lane.FillProposal(s.ctx, &proposal)

	s.Require().True(lane.blocked)

	// We expect no txs to be included in the proposal.
	s.Require().Equal(uint64(0), blockUsed.TxBytes())
	s.Require().Equal(uint64(0), blockUsed.Gas(), "0 gas")
	s.Require().Nil(iterator)

	// The proposal should contain 0 transactions in Txs().
	expectedIncludedTxs := [][]byte{}
	s.Require().Equal(0, len(proposal.txs), "no txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().Equal(lane.mempool.Select(s.ctx, nil).Tx(), tx1)

	// Calculate the remaining block space
	remainderLimit := proposal.maxBlockSpace.Sub(proposal.totalBlockSpace)

	// Call FillProposalBy with the remainder limit and iterator from the previous call.
	blockUsed = lane.FillProposalByIterator(&proposal, iterator, remainderLimit)

	// We expect no txs to be included in the proposal.
	s.Require().Equal(uint64(0), blockUsed.TxBytes())
	s.Require().Equal(uint64(0), blockUsed.Gas())

	// The proposal should contain 0 transactions in Txs().
	expectedIncludedTxs = [][]byte{}
	s.Require().Equal(0, len(proposal.txs), "no txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.txs)

	s.Require().Equal(lane.mempool.Select(s.ctx, nil).Tx(), tx1)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// createSimpleTx creates a basic single-bank-send Tx with the specified gasLimit.
func (s *LaneTestSuite) createSimpleTx(account Account, sequence uint64, gasLimit uint64) sdk.Tx {
	msg := &banktypes.MsgSend{
		FromAddress: account.Address.String(),
		ToAddress:   account.Address.String(),
	}
	txBuilder := s.encodingConfig.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		s.Require().NoError(err)
	}

	sigV2 := txsigning.SignatureV2{
		PubKey: account.PrivKey.PubKey(),
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: sequence,
	}
	err := txBuilder.SetSignatures(sigV2)
	s.Require().NoError(err)

	txBuilder.SetGasLimit(gasLimit)
	return txBuilder.GetTx()
}

// getTxBytes encodes the given transactions to raw bytes for comparison.
func (s *LaneTestSuite) getTxBytes(txs ...sdk.Tx) [][]byte {
	txBytes := make([][]byte, len(txs))
	for i, tx := range txs {
		bz, err := s.encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		txBytes[i] = bz
	}
	return txBytes
}
