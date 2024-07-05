package submitter

import (
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/bandprotocol/chain/v2/grogu/querier"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

type Submitter struct {
	contexts         []client.Context
	logger           *logger.Logger
	keyring          keyring.Keyring
	submitPriceCh    <-chan []types.SubmitPrice
	authQuerier      *querier.AuthQuerier
	txQuerier        *querier.TxQuerier
	valAddress       sdk.ValAddress
	pendingSignalIDs *sync.Map

	broadcastTimeout time.Duration
	broadcastMaxTry  uint64
	pollingInterval  time.Duration
	gasPrices        string

	idleKeyIDChannel chan string
}

func New(
	contexts []client.Context,
	logger *logger.Logger,
	keyring keyring.Keyring,
	submitPriceCh <-chan []types.SubmitPrice,
	authQuerier *querier.AuthQuerier,
	txQuerier *querier.TxQuerier,
	valAddress sdk.ValAddress,
	pendingSignalIDs *sync.Map,
	broadcastTimeout time.Duration,
	broadcastMaxTry uint64,
	pollingInterval time.Duration,
	gasPrices string,
) (*Submitter, error) {
	if len(contexts) == 0 {
		return nil, fmt.Errorf("contexts cannot be nil")
	}

	records, err := keyring.List()
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
		contexts:         contexts,
		logger:           logger,
		keyring:          keyring,
		submitPriceCh:    submitPriceCh,
		authQuerier:      authQuerier,
		txQuerier:        txQuerier,
		valAddress:       valAddress,
		pendingSignalIDs: pendingSignalIDs,
		broadcastTimeout: broadcastTimeout,
		broadcastMaxTry:  broadcastMaxTry,
		pollingInterval:  pollingInterval,
		gasPrices:        gasPrices,
		idleKeyIDChannel: idleKeyIDChannel,
	}, nil
}

func (s *Submitter) Start() {
	for {
		submitPrice := <-s.submitPriceCh
		keyID := <-s.idleKeyIDChannel
		go func(sps []types.SubmitPrice, kid string) {
			s.logger.Debug("[Submitter] starting submission")
			s.submitPrice(sps, kid)
		}(submitPrice, keyID)
	}
}

func (s *Submitter) submitPrice(prices []types.SubmitPrice, keyID string) {
	defer s.removePending(prices)
	defer func() {
		s.idleKeyIDChannel <- keyID
	}()

	msg := types.MsgSubmitPrices{
		Validator: s.valAddress.String(),
		Timestamp: time.Now().Unix(),
		Prices:    prices,
	}
	msgs := []sdk.Msg{&msg}
	memo := fmt.Sprintf("grogu:%s", version.Version)
	key, err := s.keyring.Key(keyID)
	if err != nil {
		s.logger.Error("[Submitter] failed to get key: %v", err)
		return
	}

	addr, err := key.GetAddress()
	if err != nil {
		s.logger.Error("[Submitter] failed to get key address: %v", err)
		return
	}

	gasAdjustment := 1.3
	for i := uint64(0); i < s.broadcastMaxTry; i++ {
		acc, err := s.getAccount(addr)
		if err != nil {
			s.logger.Error("[Submitter] failed to get account address: %s, with error: %v", addr.String(), err)
			return
		}

		txResp, err := broadcastMsgWithMultipleContext(s.contexts, acc, key, msgs, s.gasPrices, gasAdjustment, memo)
		if err != nil {
			s.logger.Error("[Submitter] failed to broadcast %v", err)
			continue
		}

		// if the transaction is out of gas, increase the gas adjustment
		if txResp.Codespace == sdkerrors.RootCodespace && txResp.Code == sdkerrors.ErrOutOfGas.ABCICode() {
			s.logger.Info("[Submitter] transaction is out of gas, retrying with increased gas adjustment")
			gasAdjustment += 0.1
			continue
		} else if txResp.Code != 0 {
			s.logger.Error("[Submitter] failed to broadcast with non zero code: %v", txResp.RawLog)
			continue
		}

		finalizedTxResp, err := getTxResponse(s.txQuerier, txResp.TxHash, s.broadcastTimeout, s.pollingInterval)
		if err != nil {
			s.logger.Error("[Submitter] failed to get tx response: %v", err)
			continue
		}

		switch {
		case finalizedTxResp.Code == 0:
			s.logger.Info("[Submitter] price submitted at %v", finalizedTxResp.TxHash)
			return
		case finalizedTxResp.Codespace == sdkerrors.RootCodespace && finalizedTxResp.Code == sdkerrors.ErrOutOfGas.ABCICode():
			s.logger.Info("[Submitter] transaction is out of gas, retrying with increased gas adjustment")
			gasAdjustment += 0.1
		default:
			continue
		}
	}

	s.logger.Error("[Submitter] failed to submit price")
}

func (s *Submitter) getAccount(addr sdk.AccAddress) (client.Account, error) {
	accResp, err := s.authQuerier.QueryAccount(addr)
	if err != nil {
		return nil, err
	}

	var acc client.Account
	err = unpackAccount(&acc, accResp)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *Submitter) removePending(toSubmitPrices []types.SubmitPrice) {
	for _, price := range toSubmitPrices {
		s.pendingSignalIDs.Delete(price.SignalID)
	}
}
