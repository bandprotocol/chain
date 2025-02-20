package mempool

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// -----------------------------------------------------------------------------
// Test Suite Setup
// -----------------------------------------------------------------------------

type Account struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Address sdk.AccAddress
	ConsKey cryptotypes.PrivKey
}

type MempoolTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	encodingConfig EncodingConfig
	random         *rand.Rand
	accounts       []Account
	gasTokenDenom  string
}

func TestMempoolTestSuite(t *testing.T) {
	suite.Run(t, new(MempoolTestSuite))
}

func (s *MempoolTestSuite) SetupTest() {
	s.encodingConfig = CreateTestEncodingConfig()

	s.random = rand.New(rand.NewSource(1))
	s.accounts = RandomAccounts(s.random, 5)
	s.gasTokenDenom = "uband"

	testCtx := testutil.DefaultContextWithDB(
		s.T(),
		storetypes.NewKVStoreKey("test"),
		storetypes.NewTransientStoreKey("transient_test"),
	)
	s.ctx = testCtx.Ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithBlockHeight(1)

	// Default consensus params
	s.setBlockParams(100, 1000000000000)
}

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func CreateTestEncodingConfig() EncodingConfig {
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}

func RandomAccounts(r *rand.Rand, n int) []Account {
	accs := make([]Account, n)
	for i := 0; i < n; i++ {
		pkSeed := make([]byte, 15)
		r.Read(pkSeed)

		accs[i].PrivKey = secp256k1.GenPrivKeyFromSecret(pkSeed)
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())

		accs[i].ConsKey = ed25519.GenPrivKeyFromSecret(pkSeed)
	}
	return accs
}

func (s *MempoolTestSuite) setBlockParams(maxGasLimit, maxBlockSize int64) {
	s.ctx = s.ctx.WithConsensusParams(
		tmprototypes.ConsensusParams{
			Block: &tmprototypes.BlockParams{
				MaxBytes: maxBlockSize,
				MaxGas:   maxGasLimit,
			},
		},
	)
}

// -----------------------------------------------------------------------------
// Create Mempool + Lanes
// -----------------------------------------------------------------------------

func (s *MempoolTestSuite) newMempool() *Mempool {
	signerAdapter := sdkmempool.NewDefaultSignerExtractionAdapter()

	BankSendLane := NewLane(
		log.NewTestLogger(s.T()),
		s.encodingConfig.TxConfig.TxEncoder(),
		signerAdapter,
		"bankSend",
		isBankSendTx,
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	DelegateLane := NewLane(
		log.NewTestLogger(s.T()),
		s.encodingConfig.TxConfig.TxEncoder(),
		signerAdapter,
		"delegate",
		isDelegateTx,
		math.LegacyMustNewDecFromStr("0.2"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	OtherLane := NewLane(
		log.NewTestLogger(s.T()),
		s.encodingConfig.TxConfig.TxEncoder(),
		signerAdapter,
		"other",
		isOtherTx,
		math.LegacyMustNewDecFromStr("0.4"),
		math.LegacyMustNewDecFromStr("0.4"),
		sdkmempool.DefaultPriorityMempool(),
	)

	lanes := []*Lane{BankSendLane, DelegateLane, OtherLane}

	return NewMempool(
		log.NewTestLogger(s.T()),
		lanes,
	)
}

func isBankSendTx(_ sdk.Context, tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*banktypes.MsgSend); !ok {
			return false
		}
	}
	return true
}

func isDelegateTx(_ sdk.Context, tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*stakingtypes.MsgDelegate); !ok {
			return false
		}
	}
	return true
}

func isOtherTx(_ sdk.Context, tx sdk.Tx) bool {
	// fallback if not pure bank send nor pure delegate
	return true
}

// -----------------------------------------------------------------------------
// Individual Test Methods
// -----------------------------------------------------------------------------

// TestNoTransactions ensures no transactions exist => empty proposal
func (s *MempoolTestSuite) TestNoTransactions() {
	mem := s.newMempool()

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().Equal(0, len(result.Txs))
}

// TestSingleBankTx ensures a single bank tx is included
func (s *MempoolTestSuite) TestSingleBankTx() {
	tx, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		0,
		0,
		1,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	mem := s.newMempool()
	s.Require().NoError(mem.Insert(s.ctx, tx))

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	expectedIncludedTxs := s.getTxBytes(tx)
	s.Require().Equal(1, len(result.Txs))
	s.Require().Equal(expectedIncludedTxs, result.Txs)
}

// TestOneTxPerLane checks a single transaction in each lane type
func (s *MempoolTestSuite) TestOneTxPerLane() {
	tx1, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		0,
		0,
		1,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	tx2, err := CreateDelegateTx(
		s.encodingConfig.TxConfig,
		s.accounts[1],
		0,
		0,
		1,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	tx3, err := CreateMixedTx(
		s.encodingConfig.TxConfig,
		s.accounts[2],
		0,
		0,
		1,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	mem := s.newMempool()
	// Insert in reverse order to ensure ordering is correct
	s.Require().NoError(mem.Insert(s.ctx, tx3))
	s.Require().NoError(mem.Insert(s.ctx, tx2))
	s.Require().NoError(mem.Insert(s.ctx, tx1))

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	expectedIncludedTxs := s.getTxBytes(tx1, tx2, tx3)
	s.Require().Equal(3, len(result.Txs))
	s.Require().Equal(expectedIncludedTxs, result.Txs)
}

// TestTxOverLimit checks if a tx over the block limit is rejected
func (s *MempoolTestSuite) TestTxOverLimit() {
	tx, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		0,
		0,
		101,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	mem := s.newMempool()
	s.Require().NoError(mem.Insert(s.ctx, tx))

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)

	s.Require().Equal(0, len(result.Txs))

	// Ensure the tx is removed
	for _, lane := range mem.lanes {
		s.Require().Equal(0, lane.CountTx())
	}
}

// TestTxsOverGasLimit checks if txs over the gas limit are rejected
func (s *MempoolTestSuite) TestTxsOverGasLimit() {
	bankTx1, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		0,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	bankTx2, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		1,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	delegateTx1, err := CreateDelegateTx(
		s.encodingConfig.TxConfig,
		s.accounts[1],
		0,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	delegateTx2, err := CreateDelegateTx(
		s.encodingConfig.TxConfig,
		s.accounts[1],
		1,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	otherTx1, err := CreateMixedTx(
		s.encodingConfig.TxConfig,
		s.accounts[2],
		0,
		0,
		40,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	mem := s.newMempool()
	// Insert in reverse order to ensure ordering is correct
	s.Require().NoError(mem.Insert(s.ctx, otherTx1))
	s.Require().NoError(mem.Insert(s.ctx, delegateTx2))
	s.Require().NoError(mem.Insert(s.ctx, delegateTx1))
	s.Require().NoError(mem.Insert(s.ctx, bankTx2))
	s.Require().NoError(mem.Insert(s.ctx, bankTx1))

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	// should not contain the otherTx1
	expectedIncludedTxs := s.getTxBytes(bankTx1, bankTx2, delegateTx1, delegateTx2)
	s.Require().Equal(4, len(result.Txs))
	s.Require().Equal(expectedIncludedTxs, result.Txs)
}

// TestFillUpLeftOverSpace checks if the proposal fills up the remaining space
func (s *MempoolTestSuite) TestFillUpLeftOverSpace() {
	bankTx1, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		0,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	bankTx2, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		1,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	bankTx3, err := CreateBankSendTx(
		s.encodingConfig.TxConfig,
		s.accounts[0],
		2,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	delegateTx1, err := CreateDelegateTx(
		s.encodingConfig.TxConfig,
		s.accounts[1],
		0,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	delegateTx2, err := CreateDelegateTx(
		s.encodingConfig.TxConfig,
		s.accounts[1],
		1,
		0,
		20,
		sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
	)
	s.Require().NoError(err)

	mem := s.newMempool()
	// Insert in reverse order to ensure ordering is correct
	s.Require().NoError(mem.Insert(s.ctx, delegateTx2))
	s.Require().NoError(mem.Insert(s.ctx, delegateTx1))
	s.Require().NoError(mem.Insert(s.ctx, bankTx3))
	s.Require().NoError(mem.Insert(s.ctx, bankTx2))
	s.Require().NoError(mem.Insert(s.ctx, bankTx1))

	proposal := NewProposal(
		log.NewTestLogger(s.T()),
		s.ctx.ConsensusParams().Block.MaxBytes,
		uint64(s.ctx.ConsensusParams().Block.MaxGas),
	)

	result, err := mem.PrepareProposal(s.ctx, proposal)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	// should contain bankTx3 as the last tx
	expectedIncludedTxs := s.getTxBytes(bankTx1, bankTx2, delegateTx1, delegateTx2, bankTx3)
	s.Require().Equal(5, len(result.Txs))
	s.Require().Equal(expectedIncludedTxs, result.Txs)
}

// -----------------------------------------------------------------------------
// Tx creation helpers
// -----------------------------------------------------------------------------

func CreateBankSendTx(
	txCfg client.TxConfig,
	account Account,
	nonce, timeout uint64,
	gasLimit uint64,
	fees ...sdk.Coin,
) (authsigning.Tx, error) {
	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: account.Address.String(),
			ToAddress:   account.Address.String(),
		},
	}
	return buildTx(txCfg, account, msgs, nonce, timeout, gasLimit, fees...)
}

func CreateDelegateTx(
	txCfg client.TxConfig,
	account Account,
	nonce, timeout uint64,
	gasLimit uint64,
	fees ...sdk.Coin,
) (authsigning.Tx, error) {
	msgs := []sdk.Msg{
		&stakingtypes.MsgDelegate{
			DelegatorAddress: account.Address.String(),
			ValidatorAddress: account.Address.String(),
		},
	}
	return buildTx(txCfg, account, msgs, nonce, timeout, gasLimit, fees...)
}

// MixedTx includes both a bank send and delegate to ensure it goes to "other".
func CreateMixedTx(
	txCfg client.TxConfig,
	account Account,
	nonce, timeout uint64,
	gasLimit uint64,
	fees ...sdk.Coin,
) (authsigning.Tx, error) {
	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: account.Address.String(),
			ToAddress:   account.Address.String(),
		},
		&stakingtypes.MsgDelegate{
			DelegatorAddress: account.Address.String(),
			ValidatorAddress: account.Address.String(),
		},
	}
	return buildTx(txCfg, account, msgs, nonce, timeout, gasLimit, fees...)
}

func buildTx(
	txCfg client.TxConfig,
	account Account,
	msgs []sdk.Msg,
	nonce, timeout, gasLimit uint64,
	fees ...sdk.Coin,
) (authsigning.Tx, error) {
	txBuilder := txCfg.NewTxBuilder()
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	sigV2 := txsigning.SignatureV2{
		PubKey: account.PrivKey.PubKey(),
		Data: &txsigning.SingleSignatureData{
			SignMode:  txsigning.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
		Sequence: nonce,
	}
	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	txBuilder.SetTimeoutHeight(timeout)
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(gasLimit)

	return txBuilder.GetTx(), nil
}

// getTxBytes encodes the given transactions to raw bytes for comparison.
func (s *MempoolTestSuite) getTxBytes(txs ...sdk.Tx) [][]byte {
	txBytes := make([][]byte, len(txs))
	for i, tx := range txs {
		bz, err := s.encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		txBytes[i] = bz
	}
	return txBytes
}

// decodeTxs decodes the given TxWithInfo slice back into sdk.Tx for easy comparison.
func (s *MempoolTestSuite) decodeTxs(infos []TxWithInfo) []sdk.Tx {
	res := make([]sdk.Tx, len(infos))
	for i, info := range infos {
		tx, err := s.encodingConfig.TxConfig.TxDecoder()(info.TxBytes)
		s.Require().NoError(err)
		res[i] = tx
	}
	return res
}

// extractTxBytes is a convenience to get a [][]byte from []TxWithInfo.
func extractTxBytes(txs []TxWithInfo) [][]byte {
	bz := make([][]byte, len(txs))
	for i, tx := range txs {
		bz[i] = tx.TxBytes
	}
	return bz
}
