package round1

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Round1 is a worker responsible for round1 in the DKG process of TSS module
type Round1 struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round1{}

// New creates a new instance of the Round1 worker.
// It initializes the necessary components and returns the created Round1 instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Round1, error) {
	// create http client
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round1{
		context: ctx,
		logger:  ctx.Logger.With("worker", "round1"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round1 events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round1) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round1",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS AND %s.%s = '%s'",
			types.EventTypeCreateGroup,
			types.AttributeKeyGroupID,
			types.EventTypeCreateGroup,
			types.AttributeKeyMember,
			r.context.Config.Granter,
		),
		1000,
	)
	return err
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round1) handleTxResult(txResult abci.TxResult) {
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
func (r *Round1) handleEvent(event *Event) {
	logger := r.logger.With("gid", event.GroupID)
	logger.Info(":delivery_truck: Processing incoming group event")

	mid, err := event.getMemberID(r.context.Config.Granter)
	if err != nil {
		logger.Error(":cold_sweat: Failed to get member id: %s", err.Error())
		return
	}

	data, err := tss.GenerateRound1Data(mid, event.Threshold, event.DKGContext)
	if err != nil {
		logger.Error(":cold_sweat: Failed to generate round1 data with error: %s", err.Error())
		return
	}

	// Set group data
	group := store.Group{
		MemberID:       mid,
		Coefficients:   data.Coefficients,
		OneTimePrivKey: data.OneTimePrivKey,
	}
	r.context.Store.SetGroup(event.GroupID, group)

	// Generate message
	msg := &types.MsgSubmitDKGRound1{
		GroupID:            event.GroupID,
		CoefficientsCommit: data.CoefficientsCommit,
		OneTimePubKey:      data.OneTimePubKey,
		A0Sig:              data.A0Sig,
		OneTimeSig:         data.OneTimeSig,
		Member:             r.context.Config.Granter,
	}

	// Send the message to the message channel
	r.context.MsgCh <- msg
}

// Start starts the Round1 worker.
// It subscribes to the round1 events, and continuously processes incoming events by calling handleTxResult.
func (r *Round1) Start() {
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

// Stop stops the Round1 worker.
func (r *Round1) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
