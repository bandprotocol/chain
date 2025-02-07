package mempool

import (
	"context"
	"math/rand"
	"testing"

	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/stretchr/testify/suite"

	cometabci "github.com/cometbft/cometbft/abci/types"
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
	key            *storetypes.KVStoreKey
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

	s.key = storetypes.NewKVStoreKey("test")
	testCtx := testutil.DefaultContextWithDB(s.T(), s.key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx.WithIsCheckTx(true)
	s.ctx = s.ctx.WithBlockHeight(1)
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

func (s *MempoolTestSuite) SetupSubTest() {
	s.setBlockParams(100, 1000000000000)
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

// In your actual code, you'd likely have a helper that returns a mempool with
// bank, delegate, and "other" lanes at 30%, 30%, 40%.
func (s *MempoolTestSuite) newMempool() *Mempool {
	lanes := []*Lane{
		NewLane("bankSend", isBankSendTx, 30, false),
		NewLane("delegate", isDelegateTx, 30, true),
		NewLane("other", isOtherTx, 40, false),
	}

	// Provide your actual signer extraction adapter if needed. For test, a default or mock might suffice.
	return NewMempool(s.encodingConfig.TxConfig.TxEncoder(), signerextraction.NewDefaultAdapter(), lanes)
}

func isBankSendTx(tx sdk.Tx) bool {
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

func isDelegateTx(tx sdk.Tx) bool {
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

func isOtherTx(tx sdk.Tx) bool {
	// fallback if not pure bank send nor pure delegate
	return !isBankSendTx(tx) && !isDelegateTx(tx)
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func (s *MempoolTestSuite) TestPrepareProposal() {
	s.SetupSubTest()

	s.Run("can prepare a proposal with no transactions", func() {
		mem := s.newMempool()

		// Build a "proposal handler" that uses the mempool
		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().Equal(0, len(resp.Txs))
	})

	s.Run("a single bank send tx", func() {
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
		s.Require().NoError(mem.Insert(context.Background(), tx))

		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		expected := s.getTxBytes(tx)
		s.Require().Equal(1, len(resp.Txs))
		s.Require().Equal(expected, resp.Txs)
	})

	s.Run("single tx in every type", func() {
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
		s.Require().NoError(mem.Insert(context.Background(), tx1))
		s.Require().NoError(mem.Insert(context.Background(), tx2))
		s.Require().NoError(mem.Insert(context.Background(), tx3))

		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		expected := s.getTxBytes(tx1, tx2, tx3)
		s.Require().Equal(3, len(resp.Txs))
		s.Require().Equal(expected, resp.Txs)
	})

	s.Run("one bank tx over gas limit", func() {
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
			s.accounts[1],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		delegateTx1, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[2],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		otherTx1, err := CreateMixedTx(
			s.encodingConfig.TxConfig,
			s.accounts[3],
			0,
			0,
			40,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		mem := s.newMempool()
		s.Require().NoError(mem.Insert(context.Background(), bankTx1))
		s.Require().NoError(mem.Insert(context.Background(), bankTx2))
		s.Require().NoError(mem.Insert(context.Background(), delegateTx1))
		s.Require().NoError(mem.Insert(context.Background(), otherTx1))

		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		// Expected that bankTx2 may not fit in lane's gas budget on the first pass.
		expected := s.getTxBytes(bankTx1, delegateTx1, otherTx1)
		s.Require().Equal(3, len(resp.Txs))
		s.Require().Equal(expected, resp.Txs)
	})

	s.Run("one bank tx over gas limit but has space left", func() {
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
			s.accounts[1],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		delegateTx1, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[2],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		otherTx1, err := CreateMixedTx(
			s.encodingConfig.TxConfig,
			s.accounts[3],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		mem := s.newMempool()
		s.Require().NoError(mem.Insert(context.Background(), bankTx1))
		s.Require().NoError(mem.Insert(context.Background(), bankTx2))
		s.Require().NoError(mem.Insert(context.Background(), delegateTx1))
		s.Require().NoError(mem.Insert(context.Background(), otherTx1))

		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		expected := s.getTxBytes(bankTx1, delegateTx1, otherTx1, bankTx2)
		s.Require().Equal(4, len(resp.Txs))
		s.Require().Equal(expected, resp.Txs)
	})

	s.Run("enforce one tx per signer", func() {
		delegateTx1, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		delegateTx2, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			1,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		mem := s.newMempool()
		s.Require().NoError(mem.Insert(context.Background(), delegateTx1))
		s.Require().NoError(mem.Insert(context.Background(), delegateTx2))

		ph := s.newProposalHandler(mem)

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := ph.PrepareProposalHandler()(s.ctx, &cometabci.RequestPrepareProposal{
			Height:     2,
			MaxTxBytes: maxTxBytes,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)

		expected := s.getTxBytes(delegateTx1)
		s.Require().Equal(1, len(resp.Txs))
		s.Require().Equal(expected, resp.Txs)
	})
}

// newProposalHandler is analogous to your band proposal handler that uses
// the mempool in PrepareProposal.
func (s *MempoolTestSuite) newProposalHandler(mem *Mempool) *ProposalHandler {
	// You may need to adjust: e.g. if your real code calls
	// `NewDefaultProposalHandler(logger, txDecoder, mempool)`.
	return NewDefaultProposalHandler(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxDecoder(),
		mem,
	)
}

// -----------------------------------------------------------------------------
// Tx creation helpers (same logic you had)
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

// getTxBytes encodes the given transactions.
func (s *MempoolTestSuite) getTxBytes(txs ...sdk.Tx) [][]byte {
	txBytes := make([][]byte, len(txs))
	for i, tx := range txs {
		bz, err := s.encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		txBytes[i] = bz
	}
	return txBytes
}
