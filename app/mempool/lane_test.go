package mempool

import (
	"math/rand"
	"testing"

	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/stretchr/testify/suite"

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
		signerextraction.NewDefaultAdapter(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	// Create and insert two transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 10)
	tx2 := s.createSimpleTx(s.accounts[1], 0, 10)

	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))

	// Ensure lane sees 2 transactions
	s.Require().Equal(2, lane.CountTx())
}

func (s *LaneTestSuite) TestLaneRemove() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		signerextraction.NewDefaultAdapter(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.3"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	tx := s.createSimpleTx(s.accounts[0], 0, 10)
	s.Require().NoError(lane.Insert(s.ctx, tx))
	s.Require().Equal(1, lane.CountTx())

	// Remove it
	err := lane.Remove(tx)
	s.Require().NoError(err)
	s.Require().Equal(0, lane.CountTx())
}

func (s *LaneTestSuite) TestLaneFillProposal() {
	// Lane that matches all txs
	lane := NewLane(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxEncoder(),
		signerextraction.NewDefaultAdapter(),
		"testLane",
		func(sdk.Context, sdk.Tx) bool { return true }, // accept all
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	// Insert 3 transactions
	tx1 := s.createSimpleTx(s.accounts[0], 0, 20)
	tx2 := s.createSimpleTx(s.accounts[1], 1, 20)
	tx3 := s.createSimpleTx(s.accounts[2], 2, 50) // This might be large
	tx4 := s.createSimpleTx(s.accounts[2], 3, 30) // This might be large
	tx5 := s.createSimpleTx(s.accounts[2], 4, 20)
	tx6 := s.createSimpleTx(s.accounts[2], 5, 20)
	tx7 := s.createSimpleTx(s.accounts[2], 6, 10)
	tx8 := s.createSimpleTx(s.accounts[2], 7, 10)
	s.Require().NoError(lane.Insert(s.ctx, tx1))
	s.Require().NoError(lane.Insert(s.ctx, tx2))
	s.Require().NoError(lane.Insert(s.ctx, tx3))
	s.Require().NoError(lane.Insert(s.ctx, tx4))
	s.Require().NoError(lane.Insert(s.ctx, tx5))
	s.Require().NoError(lane.Insert(s.ctx, tx6))
	s.Require().NoError(lane.Insert(s.ctx, tx7))
	s.Require().NoError(lane.Insert(s.ctx, tx8))

	// Create a proposal with block-limits
	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		1000000000000,
		100,
	)

	// FillProposal
	sizeUsed, gasUsed, iterator, txsToRemove := lane.FillProposal(s.ctx, &proposal)

	// We expect tx1 and tx2 to be included in the proposal.
	// Then the gas should be over the limit, so tx3 is yet to be considered.
	s.Require().Equal(int64(440), sizeUsed)
	s.Require().Equal(uint64(40), gasUsed, "20 gas from tx1 and 20 gas from tx2")
	s.Require().NotNil(iterator)
	s.Require().
		Len(txsToRemove, 0, "tx3 is yet to be considered")

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs := s.getTxBytes(tx1, tx2)
	s.Require().Equal(2, len(proposal.Txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.Txs)

	// Calculate the remaining block space
	remainderLimit := NewBlockSpace(proposal.Info.MaxBlockSize-sizeUsed, proposal.Info.MaxGasLimit-gasUsed)

	// Call FillProposalBy with the remainder limit and iterator from the previous call.
	sizeUsed, gasUsed, txsToRemove = lane.FillProposalBy(&proposal, iterator, remainderLimit)

	// We expect tx1, tx2, tx5, tx6, tx7, tx8 to be included in the proposal.
	s.Require().Equal(int64(884), sizeUsed)
	s.Require().Equal(uint64(60), gasUsed)
	s.Require().Equal([]sdk.Tx{tx3, tx4}, txsToRemove)
	s.Require().
		Len(txsToRemove, 2, "tx3 and tx4 are removed")

	// The proposal should contain 2 transactions in Txs().
	expectedIncludedTxs = s.getTxBytes(tx1, tx2, tx5, tx6, tx7, tx8)
	s.Require().Equal(6, len(proposal.Txs), "two txs in the proposal")
	s.Require().Equal(expectedIncludedTxs, proposal.Txs)
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
