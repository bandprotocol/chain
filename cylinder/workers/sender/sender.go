package sender

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
)

// Sender is a worker responsible for sending transactions to the node.
type Sender struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	freeKeys chan *keyring.Record
}

var _ cylinder.Worker = &Sender{}

// New creates a new instance of the Sender worker.
func New(ctx *cylinder.Context) (*Sender, error) {
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
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Sender{
		context:  ctx,
		logger:   ctx.Logger.With("worker", "sender"),
		client:   cli,
		freeKeys: freeKeys,
	}, nil
}

// Start starts the Sender worker.
func (s *Sender) Start() {
	s.logger.Info("start")

	for key := range s.freeKeys {
		msgs := s.collectMsgs()
		go s.sendMsgs(key, msgs)
	}
}

// collectMsgs collects messages from the message channel up to a limit size.
func (s *Sender) collectMsgs() []sdk.Msg {
	size := 10
	var msgs []sdk.Msg

	for len(msgs) < size && len(s.context.MsgCh) > 0 {
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

	logger := s.logger.With("msgs", GetDetail(msgs))

	// Check message validation
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			logger.Error(":exploding_head: Failed to validate basic with error: %s", err.Error())
			return
		}
	}

	logger.Info(":e-mail: Sending transaction attempt")

	res, err := s.client.BroadcastAndConfirm(key, msgs)
	if err != nil {
		logger.Error(":anxious_face_with_sweat: Cannot send messages with error: %s", err.Error())
		return
	}

	logger.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", res.TxHash)
}

// Stop stops the Sender worker.
func (s *Sender) Stop() {
	s.logger.Info("stop")
	s.client.Stop()
}
