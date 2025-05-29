package submitter

import (
	"context"
	"fmt"
	"sync"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v3/grogu/telemetry"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

type SignalPriceSubmission struct {
	SignalPrices []types.SignalPrice
	UUID         string
}

type Submitter struct {
	clientCtx           client.Context
	clients             []rpcclient.RemoteClient
	bothanClient        BothanClient
	logger              *logger.Logger
	submitSignalPriceCh <-chan SignalPriceSubmission
	authQuerier         AuthQuerier
	txQuerier           TxQuerier
	valAddress          sdk.ValAddress
	pendingSignalIDs    *sync.Map

	broadcastTimeout time.Duration
	broadcastMaxTry  uint64
	pollingInterval  time.Duration
	gasPrices        string
	gasAdjustStart   float64
	gasAdjustStep    float64

	idleKeyIDChannel chan string
}

func New(
	clientCtx client.Context,
	clients []rpcclient.RemoteClient,
	bothanClient BothanClient,
	logger *logger.Logger,
	submitSignalPriceCh <-chan SignalPriceSubmission,
	authQuerier AuthQuerier,
	txQuerier TxQuerier,
	valAddress sdk.ValAddress,
	pendingSignalIDs *sync.Map,
	broadcastTimeout time.Duration,
	broadcastMaxTry uint64,
	pollingInterval time.Duration,
	gasPrices string,
	gasAdjustStart float64,
	gasAdjustStep float64,
) (*Submitter, error) {
	if len(clients) == 0 {
		return nil, fmt.Errorf("clients cannot be empty")
	}

	records, err := clientCtx.Keyring.List()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("keyring is empty")
	}

	idleKeyIDChannel := make(chan string, len(records))
	for _, record := range records {
		idleKeyIDChannel <- record.Name
	}

	return &Submitter{
		clientCtx:           clientCtx,
		clients:             clients,
		bothanClient:        bothanClient,
		logger:              logger,
		submitSignalPriceCh: submitSignalPriceCh,
		authQuerier:         authQuerier,
		txQuerier:           txQuerier,
		valAddress:          valAddress,
		pendingSignalIDs:    pendingSignalIDs,
		broadcastTimeout:    broadcastTimeout,
		broadcastMaxTry:     broadcastMaxTry,
		pollingInterval:     pollingInterval,
		gasPrices:           gasPrices,
		gasAdjustStart:      gasAdjustStart,
		gasAdjustStep:       gasAdjustStep,
		idleKeyIDChannel:    idleKeyIDChannel,
	}, nil
}

func (s *Submitter) Start() {
	for {
		priceSubmission := <-s.submitSignalPriceCh
		since := time.Now()
		keyID := <-s.idleKeyIDChannel
		telemetry.ObserveWaitingSenderDuration(time.Since(since).Seconds())

		go func(sps SignalPriceSubmission, kid string) {
			s.logger.Debug("[Submitter] starting submission")
			s.submitPrice(sps, kid)
		}(priceSubmission, keyID)
	}
}

func (s *Submitter) submitPrice(pricesSubmission SignalPriceSubmission, keyID string) {
	telemetry.IncrementSubmittingTx()

	signalPrices, uuid := pricesSubmission.SignalPrices, pricesSubmission.UUID
	var signalIDs []string
	for _, p := range signalPrices {
		signalIDs = append(signalIDs, p.SignalID)
	}

	defer func() {
		s.removePending(signalPrices)
		s.idleKeyIDChannel <- keyID
	}()

	msg := types.MsgSubmitSignalPrices{
		Validator:    s.valAddress.String(),
		Timestamp:    time.Now().Unix(),
		SignalPrices: signalPrices,
	}
	msgs := []sdk.Msg{&msg}
	memo := fmt.Sprintf("grogu: %s, uuid: %s", version.Version, uuid)

	key, err := s.clientCtx.Keyring.Key(keyID)
	if err != nil {
		s.logger.Error("[Submitter] failed to get key: %v", err)
		telemetry.IncrementSubmitTxFailed()
		return
	}

	since := time.Now()
	gasAdjustment := s.gasAdjustStart
	for i := uint64(0); i < s.broadcastMaxTry; i++ {
		txResp, err := s.broadcastMsg(
			key,
			msgs,
			gasAdjustment,
			memo,
		)
		if err != nil {
			s.logger.Error("[Submitter] failed to broadcast: %v", err)
			continue
		}

		// if the transaction is out of gas, increase the gas adjustment
		if txResp.Codespace == sdkerrors.RootCodespace && txResp.Code == sdkerrors.ErrOutOfGas.ABCICode() {
			s.logger.Info("[Submitter] transaction is out of gas, retrying with increased gas adjustment")
			gasAdjustment += s.gasAdjustStep
			continue
		} else if txResp.Code != 0 {
			s.logger.Error("[Submitter] failed to broadcast with non zero code: %v", txResp.RawLog)
			continue
		}

		finalizedTxResp, err := s.getTxResponse(txResp.TxHash)
		if err != nil {
			s.logger.Error("[Submitter] failed to get tx response: %v", err)
			continue
		}

		switch {
		case finalizedTxResp.Code == 0:
			telemetry.ObserveSubmitTxDuration(time.Since(since).Seconds())

			s.logger.Info("[Submitter] price submitted at %v", finalizedTxResp.TxHash)
			s.pushMonitoringRecords(uuid, finalizedTxResp.TxHash, signalIDs)

			// telemetry.Set
			telemetry.ObserveSignalPriceUpdateInterval(signalPrices)
			telemetry.IncrementSubmitTxSuccess()
			return
		case finalizedTxResp.Codespace == sdkerrors.RootCodespace && finalizedTxResp.Code == sdkerrors.ErrOutOfGas.ABCICode():
			s.logger.Info("[Submitter] transaction is out of gas, retrying with increased gas adjustment")
			gasAdjustment += s.gasAdjustStep
		default:
			continue
		}
	}

	s.logger.Error("[Submitter] failed to submit price")
	telemetry.IncrementSubmitTxFailed()
}

func (s *Submitter) pushMonitoringRecords(uuid, txHash string, signalIDs []string) {
	bothanInfo, err := s.bothanClient.GetInfo()
	if err != nil {
		s.logger.Error("[Submitter] failed to query Bothan info: %v", err)
		return
	}

	if !bothanInfo.MonitoringEnabled {
		s.logger.Debug("[Submitter] monitoring is not enabled, skipping push")
		return
	}

	err = s.bothanClient.PushMonitoringRecords(uuid, txHash, signalIDs)
	if err != nil {
		s.logger.Error("[Submitter] failed to push monitoring records to Bothan: %v", err)
		return
	}

	s.logger.Info("[Submitter] successfully pushed monitoring records to Bothan")
}

func (s *Submitter) getAccountFromKey(key *keyring.Record) (client.Account, error) {
	addr, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	accResp, err := s.authQuerier.QueryAccount(addr)
	if err != nil {
		return nil, err
	}

	var acc client.Account
	registry := s.clientCtx.InterfaceRegistry
	err = registry.UnpackAny(accResp.Account, &acc)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack account with error: %v", err)
	}
	return acc, nil
}

func (s *Submitter) removePending(prices []types.SignalPrice) {
	for _, p := range prices {
		_, loaded := s.pendingSignalIDs.LoadAndDelete(p.SignalID)
		if !loaded {
			s.logger.Error("[Submitter] Attempted to delete Signal ID %s which was not pending", p.SignalID)
		}
	}
}

func (s *Submitter) broadcastMsg(
	key *keyring.Record,
	msgs []sdk.Msg,
	gasAdjustment float64,
	memo string,
) (*sdk.TxResponse, error) {
	if len(s.clients) == 0 {
		return nil, fmt.Errorf("no client provided")
	}

	txBytes, err := s.buildSignedTx(key, msgs, gasAdjustment, memo)
	if err != nil {
		return nil, err
	}

	resultsCh := make(chan *sdk.TxResponse, len(s.clients))
	failureCh := make(chan error, len(s.clients))
	for _, client := range s.clients {
		go func(client rpcclient.RemoteClient) {
			res, err := s.clientCtx.WithClient(client).BroadcastTx(txBytes)
			if err != nil {
				failureCh <- err
				return
			}
			resultsCh <- res
		}(client)
	}

	var res *sdk.TxResponse
	for range s.clients {
		select {
		case currentResult := <-resultsCh:
			if currentResult.Code == 0 {
				return currentResult, nil
			}

			res = currentResult
		case err = <-failureCh:
			continue
		}
	}

	if res != nil {
		return res, nil
	}

	return nil, err
}

func (s *Submitter) buildSignedTx(
	key *keyring.Record,
	msgs []sdk.Msg,
	gasAdjustment float64,
	memo string,
) ([]byte, error) {
	account, err := s.getAccountFromKey(key)
	if err != nil {
		return nil, err
	}

	addr, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	execMsg := authz.NewMsgExec(addr, msgs)
	gasCh := make(chan uint64, len(s.clients))
	errCh := make(chan error, len(s.clients))

	txf := tx.Factory{}.
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence()).
		WithTxConfig(s.clientCtx.TxConfig).
		WithSimulateAndExecute(true).
		WithGasAdjustment(gasAdjustment).
		WithChainID(s.clientCtx.ChainID).
		WithMemo(memo).
		WithGasPrices(s.gasPrices).
		WithKeybase(s.clientCtx.Keyring).
		WithFromName(key.Name).
		WithAccountRetriever(s.clientCtx.AccountRetriever)

	for _, client := range s.clients {
		go func(client rpcclient.RemoteClient) {
			_, adjusted, err := tx.CalculateGas(s.clientCtx.WithClient(client), txf, &execMsg)
			if err != nil {
				errCh <- err
				return
			}

			gasCh <- adjusted
		}(client)
	}

	maxGas := uint64(0)
	for range s.clients {
		select {
		case gas := <-gasCh:
			if gas > maxGas {
				maxGas = gas
			}
		case err = <-errCh:
			continue
		}
	}

	if maxGas == 0 {
		return nil, fmt.Errorf("failed to calculate gas with error: %v", err)
	}

	txf = txf.WithGas(maxGas)

	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txf, key.Name, txb, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := s.clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

func (s *Submitter) getTxResponse(
	txHash string,
) (*sdk.TxResponse, error) {
	var resp *sdk.TxResponse
	var err error

	for start := time.Now(); time.Since(start) < s.broadcastTimeout; {
		time.Sleep(s.pollingInterval)
		resp, err = s.txQuerier.QueryTx(txHash)
		if err != nil {
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("timeout exceeded with error: %v", err)
}
