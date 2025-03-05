package sender

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// Sender is a worker responsible for sending transactions to the node.
type Sender struct {
	context  *context.Context
	logger   *logger.Logger
	client   *client.Client
	freeKeys chan *keyring.Record
}

var _ cylinder.Worker = &Sender{}

// New creates a new instance of the Sender worker.
func New(ctx *context.Context) (*Sender, error) {
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
		context:  ctx,
		logger:   ctx.Logger.With("worker", "Sender"),
		client:   cli,
		freeKeys: freeKeys,
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
func (s *Sender) collectMsgs() []sdk.Msg {
	maxSize := int(s.context.Config.MaxMessages)
	var msgs []sdk.Msg

	for len(msgs) == 0 || (len(msgs) < maxSize && len(s.context.MsgCh) > 0) {
		msg := <-s.context.MsgCh
		msgs = append(msgs, msg)
	}

	return msgs
}

// sendMsgs sends the given messages using the provided key.
func (s *Sender) sendMsgs(key *keyring.Record, msgs []sdk.Msg) {
	// Return key to the free keys after function ends
	defer func() {
		s.freeKeys <- key
	}()

	// since is used to measure the time for sending messages
	since := time.Now()

	logger := s.logger.With("msgs", GetMsgDetails(msgs...))

	logger.Info(":e-mail: Sending transaction attempt")

	res, err := s.client.BroadcastAndConfirm(logger, key, msgs)
	if err != nil {
		logger.Error(":anxious_face_with_sweat: Cannot send messages with error: %s", err)

		metrics.IncSubmitTxFailedCount()
		return
	} else if res.Code != 0 {
		logger.Error(":anxious_face_with_sweat: Cannot send messages with error code: codespace: %s, code: %d", res.Codespace, res.Code)

		metrics.IncSubmitTxFailedCount()
		return
	}

	logger.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", res.TxHash)

	metrics.ObserveSubmitTxTime(time.Since(since).Seconds())
	metrics.IncSubmitTxSuccessCount()
}

// Stop stops the Sender worker.
func (s *Sender) Stop() error {
	s.logger.Info("stop")
	return s.client.Stop()
}
