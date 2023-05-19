package round2

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Round2 is a worker responsible for round2 in the DKG process of TSS module
type Round2 struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round2{}

// New creates a new instance of the Round2 worker.
// It initializes the necessary components and returns the created Round2 instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Round2, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round2{
		context: ctx,
		logger:  ctx.Logger.With("worker", "round2"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round2 events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round2) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round2",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS",
			types.EventTypeRound1Success,
			types.AttributeKeyGroupID,
		),
		1000,
	)
	return err
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round2) handleTxResult(txResult abci.TxResult) {
	msgLogs, err := event.GetMessageLogs(txResult)
	if err != nil {
		r.logger.Error("Failed to get message logs: %s", err.Error())
		return
	}

	for _, log := range msgLogs {
		event, err := ParseEvent(log)
		if err != nil {
			r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err.Error())
			return
		}

		go r.handleEvent(event)
	}
}

// handleEvent processes an incoming group event.
func (r *Round2) handleEvent(event *Event) {
	logger := r.logger.With("gid", event.GroupID)
	logger.Info(":delivery_truck: Processing incoming group event")

	// Set group data
	group, err := r.context.Store.GetGroup(event.GroupID)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err.Error())
		return
	}

	gr, err := r.client.QueryGroup(event.GroupID)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err.Error())
		return
	}

	// Get all one time public keys in the group
	oneTimePubKeys := make(tss.PublicKeys, gr.Group.Size_)
	for mid, commitment := range gr.AllRound1Commitments {
		oneTimePubKeys[mid-1] = commitment.OneTimePubKey
	}

	// Compute encrypted secret shares
	encSecretShares, err := tss.ComputeEncryptedSecretShares(
		group.MemberID,
		group.OneTimePrivKey,
		oneTimePubKeys,
		group.Coefficients,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to genrate encrypted secret shares: %s", err.Error())
		return
	}

	// Generate message for round 2
	msg := &types.MsgSubmitDKGRound2{
		GroupID: event.GroupID,
		Round2Share: &types.Round2Share{
			EncryptedSecretShares: encSecretShares,
		},
		Member: r.context.Config.Granter,
	}

	r.context.MsgCh <- msg
}

// Start starts the Round2 worker.
// It subscribes to round2 events and starts processing incoming events.
func (r *Round2) Start() {
	r.logger.Info("start")

	err := r.subscribe()
	if err != nil {
		r.context.ErrCh <- err
		return
	}

	for ev := range r.eventCh {
		go r.handleTxResult(ev.Data.(tmtypes.EventDataTx).TxResult)
	}
}

// Stop stops the Round2 worker.
func (r *Round2) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
