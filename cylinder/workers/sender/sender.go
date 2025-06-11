package sender

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/cylinder/msg"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// Sender is a worker responsible for sending transactions to the node.
type Sender struct {
	context   *context.Context
	logger    *logger.Logger
	client    *client.Client
	freeKeys  chan *keyring.Record
	receivers []*msg.ResponseReceiver
}

var _ cylinder.Worker = &Sender{}

// New creates a new instance of the Sender worker.
func New(ctx *context.Context, receivers []*msg.ResponseReceiver) (*Sender, error) {
	// Add all keys to free keys
	keys, err := ctx.Keyring.List()
	if err != nil {
		return nil, err
	}

	freeKeys := make(chan *keyring.Record, len(keys))
	for _, key := range keys {
		freeKeys <- key
	}

	// Create a client
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &Sender{
		context:   ctx,
		logger:    ctx.Logger.With("worker", "Sender"),
		client:    cli,
		freeKeys:  freeKeys,
		receivers: receivers,
	}, nil
}

// Start starts the Sender worker.
func (s *Sender) Start() {
	s.logger.Info("start")

	for {
		// since is used to measure the time for waiting for free keys
		since := time.Now()

		key := <-s.freeKeys
		metrics.ObserveWaitingSenderTime(time.Since(since).Seconds())

		msgs := s.collectMsgs()
		go s.sendMsgs(key, msgs)

		metrics.AddSubmittingTxCount(float64(len(msgs)))
	}
}

// collectMsgs collects messages from the message channel up to a limit size.
// It will block until got at least one message, then return non-empty message list.
func (s *Sender) collectMsgs() []msg.Request {
	maxSize := int(s.context.Config.MaxMessages)
	var msgs []msg.Request

	// drain first message (priority channel first)
	select {
	case msg := <-s.context.PriorityMsgRequestCh:
		msgs = append(msgs, msg)
	default:
		select {
		case msg := <-s.context.PriorityMsgRequestCh:
			msgs = append(msgs, msg)
		case msg := <-s.context.MsgRequestCh:
			msgs = append(msgs, msg)
		}
	}

	// wait for 0.1 second to collect more messages.
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	for len(msgs) < maxSize {
		select {
		case msg := <-s.context.PriorityMsgRequestCh:
			msgs = append(msgs, msg)
		case <-timer.C:
			return msgs
		default:
			select {
			case msg := <-s.context.PriorityMsgRequestCh:
				msgs = append(msgs, msg)
			case msg := <-s.context.MsgRequestCh:
				msgs = append(msgs, msg)
			case <-timer.C:
				return msgs
			}
		}
	}

	return msgs
}

// sendMsgs sends the given messages using the provided key.
func (s *Sender) sendMsgs(key *keyring.Record, msgs []msg.Request) {
	// Return key to the free keys after function ends
	defer func() {
		s.freeKeys <- key
	}()

	// since is used to measure the time for sending messages
	since := time.Now()

	sdkMsgs := make([]sdk.Msg, len(msgs))
	for i, msg := range msgs {
		sdkMsgs[i] = msg.Msg
	}

	logger := s.logger.With("msgs", GetMsgDetails(sdkMsgs...))
	logger.Info(":e-mail: Sending transaction attempt")

	res, err := s.client.BroadcastAndConfirm(logger, key, sdkMsgs)
	if err != nil {
		logger.Error(":anxious_face_with_sweat: Cannot send messages with error: %s", err)

		metrics.IncSubmitTxFailedCount()
		s.retryMsgs(msgs, err)
		return
	} else if res.Code != 0 {
		logger.Error(":anxious_face_with_sweat: Cannot send messages with error code: codespace: %s, code: %d", res.Codespace, res.Code)

		metrics.IncSubmitTxFailedCount()
		s.retryMsgs(msgs, fmt.Errorf("error with codespace: %s, code: %d", res.Codespace, res.Code))
		return
	}

	logger.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", res.TxHash)

	metrics.ObserveSubmitTxTime(time.Since(since).Seconds())
	metrics.IncSubmitTxSuccessCount()

	s.forwardResult(msgs, true, res.TxHash, nil)
}

// retryMsgs retries the messages that failed to send, but if the request
// reaches max retry, it will forward the result to the receiver.
func (s *Sender) retryMsgs(msgs []msg.Request, err error) {
	var reachedMaxRetryMsgs []msg.Request

	for _, m := range msgs {
		msgLog := s.logger.With("msg", GetMsgDetails([]sdk.Msg{m.Msg}...))
		if m.Retry < m.MaxRetry {
			msgLog.Warn(
				":anxious_face_with_sweat: Failed to send ID: %d, retry: %d; %s",
				m.ID,
				m.Retry,
				err,
			)

			s.context.MsgRequestCh <- m.IncreaseRetry()
		} else {
			msgLog.Error(
				":anxious_face_with_sweat: Failed to send request ID: %d, retry: %d; %s",
				m.ID,
				m.Retry,
				err,
			)

			reachedMaxRetryMsgs = append(reachedMaxRetryMsgs, m)
		}
	}

	s.forwardResult(reachedMaxRetryMsgs, false, "", err)
}

// forwardResult forwards the result of the message to the receiver.
func (s *Sender) forwardResult(msgs []msg.Request, success bool, txHash string, err error) {
	for _, m := range msgs {
		for _, receiver := range s.receivers {
			if receiver.ReqType == m.ReqType {
				receiver.ResponseCh <- msg.NewResponse(m, success, txHash, err)
			}
		}
	}
}

// Stop stops the Sender worker.
func (s *Sender) Stop() error {
	s.logger.Info("stop")
	return s.client.Stop()
}

// GetResponseReceivers returns the message response receivers of the worker.
func (s *Sender) GetResponseReceivers() []*msg.ResponseReceiver {
	return s.receivers
}
