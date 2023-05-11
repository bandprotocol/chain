package sender

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
)

type Sender struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	freeKeys chan *keyring.Record
}

var _ cylinder.Worker = &Sender{}

func New(ctx *cylinder.Context) (*Sender, error) {
	// add all keys to free keys
	keys, err := ctx.Keyring.List()
	if err != nil {
		return nil, err
	}
	freeKeys := make(chan *keyring.Record, len(keys))
	for _, key := range keys {
		freeKeys <- key
	}

	// create http client
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

func (s *Sender) Start() {
	s.logger.Info("start")

	for {
		key := <-s.freeKeys

		// get at most 10 messages from Msg channel to prevent too big transactions
		size := 10
		var msgs []sdk.Msg
		for {
			// break the look to send messages if:
			// - collected messages are more than limit size
			// - no message left in the channel
			if len(msgs) >= size || (len(msgs) > 0 && len(s.context.MsgCh) == 0) {
				break
			}
			msg := <-s.context.MsgCh
			msgs = append(msgs, msg)
		}

		// send messages
		go s.sendMsgs(key, msgs)
	}
}

func (s *Sender) sendMsgs(key *keyring.Record, msgs []sdk.Msg) {
	// Return key and update pending metric when done with SubmitReport whether successfully or not.
	defer func() {
		s.freeKeys <- key
	}()

	logger := s.logger.With("msgs", GetDetail(msgs))

	// check message validation
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
	return
}

func (s *Sender) Stop() {
	s.logger.Info("stop")
	s.client.Stop()
}
