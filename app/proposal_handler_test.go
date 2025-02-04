package band

import (
	"math/rand"
	"testing"

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
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	cometabci "github.com/cometbft/cometbft/abci/types"
	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

type Account struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Address sdk.AccAddress
	ConsKey cryptotypes.PrivKey
}

type ProposalHandlerTestSuite struct {
	suite.Suite
	ctx sdk.Context
	key *storetypes.KVStoreKey

	encodingConfig EncodingConfig
	random         *rand.Rand
	accounts       []Account
	gasTokenDenom  string
}

func TestProposalHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalHandlerTestSuite))
}

func (s *ProposalHandlerTestSuite) SetupTest() {
	// Set up basic TX encoding config.
	s.encodingConfig = CreateTestEncodingConfig()

	// Create a few random accounts
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

func (s *ProposalHandlerTestSuite) SetupSubTest() {
	s.setBlockParams(100, 1000000000000)
}

func (s *ProposalHandlerTestSuite) setBlockParams(maxGasLimit, maxBlockSize int64) {
	s.ctx = s.ctx.WithConsensusParams(
		tmprototypes.ConsensusParams{
			Block: &tmprototypes.BlockParams{
				MaxBytes: maxBlockSize,
				MaxGas:   maxGasLimit,
			},
		},
	)
}

func (s *ProposalHandlerTestSuite) setUpProposalHandlers(bandMempool *BandMempool) *ProposalHandler {
	return NewDefaultProposalHandler(
		log.NewNopLogger(),
		s.encodingConfig.TxConfig.TxDecoder(),
		bandMempool,
	)
}

func (s *ProposalHandlerTestSuite) TestPrepareProposal() {
	s.Run("can prepare a proposal with no transactions", func() {
		// Create a new BandMempool
		bandMempool := NewBandMempool(s.encodingConfig.TxConfig.TxEncoder())

		// Set up the band proposal handler with no transactions
		proposalHandler := s.setUpProposalHandlers(bandMempool).PrepareProposalHandler()

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := proposalHandler(s.ctx, &cometabci.RequestPrepareProposal{Height: 2, MaxTxBytes: maxTxBytes})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
		s.Require().Equal(0, len(resp.Txs))
	})

	s.Run("can build a proposal with a single bank send tx", func() {
		// Create a bank transaction that will be inserted into the bank lane
		tx, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a new BandMempool
		bandMempool := NewBandMempool(s.encodingConfig.TxConfig.TxEncoder())

		// Insert the transaction into the mempool
		err = bandMempool.Insert(s.ctx, tx)
		s.Require().NoError(err)

		// Set up the band proposal handler with no transactions
		proposalHandler := s.setUpProposalHandlers(bandMempool).PrepareProposalHandler()

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := proposalHandler(s.ctx, &cometabci.RequestPrepareProposal{Height: 2, MaxTxBytes: maxTxBytes})
		s.Require().NotNil(resp)
		s.Require().NoError(err)

		proposal := s.getTxBytes(tx)
		s.Require().Equal(1, len(resp.Txs))
		s.Require().Equal(proposal, resp.Txs)
	})

	s.Run("can build a proposal with single tx in every types", func() {
		// Create a bank transaction that will be inserted into the bank lane
		tx1, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a delegate transaction that will be inserted into the delegate lane
		tx2, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[1],
			0,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a mixed transaction that will be inserted into the other lane
		tx3, err := CreateMixedTx(
			s.encodingConfig.TxConfig,
			s.accounts[2],
			0,
			0,
			1,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a new BandMempool
		bandMempool := NewBandMempool(s.encodingConfig.TxConfig.TxEncoder())

		// Insert the transaction into the mempool
		err = bandMempool.Insert(s.ctx, tx1)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, tx2)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, tx3)
		s.Require().NoError(err)

		// Set up the band proposal handler with no transactions
		proposalHandler := s.setUpProposalHandlers(bandMempool).PrepareProposalHandler()

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := proposalHandler(s.ctx, &cometabci.RequestPrepareProposal{Height: 2, MaxTxBytes: maxTxBytes})
		s.Require().NotNil(resp)
		s.Require().NoError(err)

		proposal := s.getTxBytes(tx1, tx2, tx3)
		s.Require().Equal(3, len(resp.Txs))
		s.Require().Equal(proposal, resp.Txs)
	})

	s.Run("can build a proposal with one bank tx over gas limit", func() {
		// Create a bank transaction that will be inserted into the bank lane
		bankTx1, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a bank transaction that will be inserted into the bank lane but over gas limit
		bankTx2, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[1],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a delegate transaction that will be inserted into the delegate lane
		delegateTx1, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[2],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a mixed transaction that will be inserted into the other lane
		otherTx1, err := CreateMixedTx(
			s.encodingConfig.TxConfig,
			s.accounts[3],
			0,
			0,
			40,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a new BandMempool
		bandMempool := NewBandMempool(s.encodingConfig.TxConfig.TxEncoder())

		// Insert the transaction into the mempool
		err = bandMempool.Insert(s.ctx, bankTx1)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, bankTx2)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, delegateTx1)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, otherTx1)
		s.Require().NoError(err)

		// Set up the band proposal handler with no transactions
		proposalHandler := s.setUpProposalHandlers(bandMempool).PrepareProposalHandler()

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := proposalHandler(s.ctx, &cometabci.RequestPrepareProposal{Height: 2, MaxTxBytes: maxTxBytes})
		s.Require().NotNil(resp)
		s.Require().NoError(err)

		proposal := s.getTxBytes(bankTx1, delegateTx1, otherTx1)
		s.Require().Equal(3, len(resp.Txs))
		s.Require().Equal(proposal, resp.Txs)
	})

	s.Run("can build a proposal with one bank tx over gas limit but has space left", func() {
		// Create a bank transaction that will be inserted into the bank lane
		bankTx1, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[0],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a bank transaction that will be inserted into the bank lane but over gas limit
		bankTx2, err := CreateBankSendTx(
			s.encodingConfig.TxConfig,
			s.accounts[1],
			0,
			0,
			20,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a delegate transaction that will be inserted into the delegate lane
		delegateTx1, err := CreateDelegateTx(
			s.encodingConfig.TxConfig,
			s.accounts[2],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a mixed transaction that will be inserted into the other lane
		otherTx1, err := CreateMixedTx(
			s.encodingConfig.TxConfig,
			s.accounts[3],
			0,
			0,
			30,
			sdk.NewCoin(s.gasTokenDenom, math.NewInt(1000000)),
		)
		s.Require().NoError(err)

		// Create a new BandMempool
		bandMempool := NewBandMempool(s.encodingConfig.TxConfig.TxEncoder())

		// Insert the transaction into the mempool
		err = bandMempool.Insert(s.ctx, bankTx1)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, bankTx2)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, delegateTx1)
		s.Require().NoError(err)

		err = bandMempool.Insert(s.ctx, otherTx1)
		s.Require().NoError(err)

		// Set up the band proposal handler with no transactions
		proposalHandler := s.setUpProposalHandlers(bandMempool).PrepareProposalHandler()

		maxTxBytes := s.ctx.ConsensusParams().Block.MaxBytes
		resp, err := proposalHandler(s.ctx, &cometabci.RequestPrepareProposal{Height: 2, MaxTxBytes: maxTxBytes})
		s.Require().NotNil(resp)
		s.Require().NoError(err)

		proposal := s.getTxBytes(bankTx1, delegateTx1, otherTx1, bankTx2)
		s.Require().Equal(4, len(resp.Txs))
		s.Require().Equal(proposal, resp.Txs)
	})
}

func CreateBankSendTx(txCfg client.TxConfig, account Account, nonce, timeout uint64, gasLimit uint64, fees ...sdk.Coin) (authsigning.Tx, error) {
	msgs := make([]sdk.Msg, 1)
	msgs[0] = &banktypes.MsgSend{
		FromAddress: account.Address.String(),
		ToAddress:   account.Address.String(),
	}

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

func CreateDelegateTx(txCfg client.TxConfig, account Account, nonce, timeout uint64, gasLimit uint64, fees ...sdk.Coin) (authsigning.Tx, error) {
	msgs := make([]sdk.Msg, 1)
	msgs[0] = &stakingtypes.MsgDelegate{
		DelegatorAddress: account.Address.String(),
		ValidatorAddress: account.Address.String(),
	}

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

func CreateMixedTx(txCfg client.TxConfig, account Account, nonce, timeout uint64, gasLimit uint64, fees ...sdk.Coin) (authsigning.Tx, error) {
	msgs := make([]sdk.Msg, 2)
	msgs[0] = &banktypes.MsgSend{
		FromAddress: account.Address.String(),
		ToAddress:   account.Address.String(),
	}
	msgs[1] = &stakingtypes.MsgDelegate{
		DelegatorAddress: account.Address.String(),
		ValidatorAddress: account.Address.String(),
	}

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

func (s *ProposalHandlerTestSuite) getTxBytes(txs ...sdk.Tx) [][]byte {
	txBytes := make([][]byte, len(txs))
	for i, tx := range txs {
		bz, err := s.encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)

		txBytes[i] = bz
	}
	return txBytes
}
