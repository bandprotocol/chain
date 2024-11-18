package submitter

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/bothan/v1"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/grogu/submitter/testutil"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

type SubmitterTestSuite struct {
	suite.Suite

	Submitter           *Submitter
	SubmitSignalPriceCh chan SignalPriceSubmission
}

func TestSubmitterTestSuite(t *testing.T) {
	suite.Run(t, new(SubmitterTestSuite))
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", ".band")
	if err != nil {
		dir = band.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}

func (s *SubmitterTestSuite) SetupTest() {
	// Initialize encoding config
	initAppOptions := viper.New()
	tempDir := tempDir()
	initAppOptions.Set(flags.FlagHome, tempDir)
	tempApplication := band.NewBandApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tempDir,
		initAppOptions,
		[]wasmkeeper.Option{},
		100,
	)

	// Setup keyring
	cdc := tempApplication.AppCodec()
	kb := keyring.NewInMemory(cdc)
	_, _, err := kb.NewMnemonic(
		"test",
		keyring.English,
		sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	s.Require().NoError(err)

	// Setup Client Context
	clientCtx := client.Context{
		ChainID:           "mock-chain",
		Codec:             cdc,
		InterfaceRegistry: tempApplication.InterfaceRegistry(),
		Keyring:           kb,
		TxConfig:          tempApplication.GetTxConfig(),
		BroadcastMode:     flags.BroadcastSync,
	}

	ctrl := gomock.NewController(s.T())
	mockClient := testutil.NewMockRemoteClient(ctrl)
	mockClient.EXPECT().Remote().Return("mock").AnyTimes()
	mockClient.EXPECT().
		ABCIQueryWithOptions(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, data bytes.HexBytes, opts rpcclient.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
			gInfo := sdk.GasInfo{
				GasWanted: 100,
				GasUsed:   100,
			}
			simRes := &sdk.SimulationResponse{
				GasInfo: gInfo,
				Result:  nil,
			}

			bz, _ := codec.NewProtoCodec(tempApplication.InterfaceRegistry()).GRPCCodec().Marshal(simRes)

			return &coretypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Codespace: sdkerrors.RootCodespace,
					Height:    1,
					Value:     bz,
				},
			}, nil
		}).
		AnyTimes()

	mockRPCClients := []rpcclient.RemoteClient{mockClient}

	mockBothanClient := testutil.NewMockBothanClient(ctrl)
	mockBothanClient.EXPECT().GetInfo().Return(&bothan.GetInfoResponse{MonitoringEnabled: true}, nil).AnyTimes()
	mockBothanClient.EXPECT().PushMonitoringRecords(gomock.Any(), gomock.Any()).AnyTimes()

	mockAuthQuerier := testutil.NewMockAuthQuerier(ctrl)
	mockAuthQuerier.EXPECT().
		QueryAccount(gomock.Any()).
		DoAndReturn(func(address sdk.Address) (*auth.QueryAccountResponse, error) {
			account := auth.NewBaseAccountWithAddress(sdk.MustAccAddressFromBech32(address.String()))
			any, _ := codectypes.NewAnyWithValue(account)
			return &auth.QueryAccountResponse{Account: any}, nil
		}).
		AnyTimes()

	mockTxQuerier := testutil.NewMockTxQuerier(ctrl)
	mockTxQuerier.EXPECT().
		QueryTx(gomock.Any()).
		Return(&sdk.TxResponse{TxHash: "mock-tx-hash", Code: 0}, nil).
		AnyTimes()

	// Initialize logger
	allowLevel, _ := log.ParseLogLevel("info")
	l := logger.NewLogger(allowLevel)

	// Create submit channel
	submitSignalPriceCh := make(chan SignalPriceSubmission, 300)

	// Set up validator address
	validAddress := sdk.ValAddress("1000000001")

	// Initialize pending signal IDs map
	pendingSignalIDs := sync.Map{}

	// Create submitter instance
	submitterInstance, err := New(
		clientCtx,
		mockRPCClients,
		mockBothanClient,
		l,
		submitSignalPriceCh,
		mockAuthQuerier,
		mockTxQuerier,
		validAddress,
		&pendingSignalIDs,
		10*time.Second,
		3,
		1*time.Second,
		"0.025stake",
	)
	s.Require().NoError(err)
	s.Submitter = submitterInstance
	s.SubmitSignalPriceCh = submitSignalPriceCh
}

func (s *SubmitterTestSuite) TestSubmitterSubmitPrice() {
	// Override the BroadcastTx function to simulate out of gas error
	mockClient := s.Submitter.clients[0].(*testutil.MockRemoteClient)
	mockClient.EXPECT().
		BroadcastTxAsync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: 0}, nil).
		AnyTimes()
	mockClient.EXPECT().
		BroadcastTxSync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: 0}, nil).
		AnyTimes()

	s.Submitter.clients = []rpcclient.RemoteClient{mockClient}

	// Add signal price data to channel
	prices := []types.SignalPrice{
		{
			SignalID: "signal1",
			Price:    12345,
			Status:   types.SIGNAL_PRICE_STATUS_AVAILABLE,
		},
	}

	signalPriceSubmission := SignalPriceSubmission{
		SignalPrices: prices,
		UUID:         "uuid1",
	}
	s.SubmitSignalPriceCh <- signalPriceSubmission
	s.Submitter.pendingSignalIDs.Store("signal1", struct{}{})

	// Check length of idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 1)

	// Get key ID from idleKeyIDChannel
	keyID := <-s.Submitter.idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 0)

	s.Submitter.submitPrice(signalPriceSubmission, keyID)

	// Check pending signal IDs
	_, pending := s.Submitter.pendingSignalIDs.Load("signal1")
	s.Require().False(pending, "Signal ID should have been removed from pendingSignalIDs")

	// Check key ID added back to idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 1)
}

func (s *SubmitterTestSuite) TestSubmitterSubmitPrice_OutOfGas() {
	// Override the BroadcastTx function to simulate out of gas error
	mockClient := s.Submitter.clients[0].(*testutil.MockRemoteClient)
	mockClient.EXPECT().
		BroadcastTxAsync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: sdkerrors.ErrOutOfGas.ABCICode()}, nil).
		AnyTimes()
	mockClient.EXPECT().
		BroadcastTxSync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: sdkerrors.ErrOutOfGas.ABCICode()}, nil).
		AnyTimes()

	s.Submitter.clients = []rpcclient.RemoteClient{mockClient}

	// Add signal price data to channel
	prices := []types.SignalPrice{
		{
			SignalID: "signal1",
			Price:    12345,
			Status:   types.SIGNAL_PRICE_STATUS_AVAILABLE,
		},
	}

	signalPriceSubmission := SignalPriceSubmission{
		SignalPrices: prices,
		UUID:         "uuid1",
	}
	s.SubmitSignalPriceCh <- signalPriceSubmission
	s.Submitter.pendingSignalIDs.Store("signal1", struct{}{})

	// Check length of idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 1)

	// Get key ID from idleKeyIDChannel
	keyID := <-s.Submitter.idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 0)

	s.Submitter.submitPrice(signalPriceSubmission, keyID)

	// Check pending signal IDs
	_, pending := s.Submitter.pendingSignalIDs.Load("signal1")
	s.Require().False(pending, "Signal ID should have been removed from pendingSignalIDs")

	// Check key ID added back to idleKeyIDChannel
	s.Require().Len(s.Submitter.idleKeyIDChannel, 1)
}

func (s *SubmitterTestSuite) TestSubmitterBuildSignedTx() {
	keyID := <-s.Submitter.idleKeyIDChannel
	key, err := s.Submitter.clientCtx.Keyring.Key(keyID)
	s.Require().NoError(err)

	msg := types.MsgSubmitSignalPrices{
		Validator: s.Submitter.valAddress.String(),
		Timestamp: time.Now().Unix(),
		SignalPrices: []types.SignalPrice{
			{
				SignalID: "signal1",
				Price:    12345,
				Status:   types.SIGNAL_PRICE_STATUS_AVAILABLE,
			},
		},
	}
	msgs := []sdk.Msg{&msg}

	txBytes, err := s.Submitter.buildSignedTx(key, msgs, 1.3, "test-memo")
	s.Require().NoError(err)
	s.Require().NotNil(txBytes)
}

func (s *SubmitterTestSuite) TestSubmitterBroadcastMsg() {
	// Override the BroadcastTx function to simulate out of gas error
	mockClient := s.Submitter.clients[0].(*testutil.MockRemoteClient)
	mockClient.EXPECT().
		BroadcastTxAsync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: 0}, nil).
		AnyTimes()
	mockClient.EXPECT().
		BroadcastTxSync(gomock.Any(), gomock.Any()).
		Return(&coretypes.ResultBroadcastTx{Code: 0}, nil).
		AnyTimes()

	keyID := <-s.Submitter.idleKeyIDChannel
	key, err := s.Submitter.clientCtx.Keyring.Key(keyID)
	s.Require().NoError(err)

	msg := types.MsgSubmitSignalPrices{
		Validator: s.Submitter.valAddress.String(),
		Timestamp: time.Now().Unix(),
		SignalPrices: []types.SignalPrice{
			{
				SignalID: "signal1",
				Price:    12345,
				Status:   types.SIGNAL_PRICE_STATUS_AVAILABLE,
			},
		},
	}
	msgs := []sdk.Msg{&msg}

	_, err = s.Submitter.broadcastMsg(key, msgs, 1.3, "test-memo")
	s.Require().NoError(err)
}

func (s *SubmitterTestSuite) TestSubmitterGetTxResponse() {
	// Simulate a successful transaction
	txHash := "mock-tx-hash"
	resp, err := s.Submitter.getTxResponse(txHash)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(txHash, resp.TxHash)
	s.Require().Equal(uint32(0), resp.Code)
}
