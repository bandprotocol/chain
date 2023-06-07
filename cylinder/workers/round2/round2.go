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

// subscribe subscribes to the round1_success events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round2) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s EXISTS",
		types.EventTypeRound1Success,
		types.AttributeKeyGroupID,
	)
	r.eventCh, err = r.client.Subscribe("Round2", subscriptionQuery, 1000)
	return
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round2) handleTxResult(txResult abci.TxResult) {
	msgLogs, err := event.GetMessageLogs(txResult)
	if err != nil {
		r.logger.Error("Failed to get message logs: %s", err)
		return
	}

	for _, log := range msgLogs {
		event, err := ParseEvent(log)
		if err != nil {
			r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
			return
		}

		go r.handleGroup(event.GroupID)
	}
}

// handleGroup processes an incoming group.
func (r *Round2) handleGroup(gid tss.GroupID) {
	logger := r.logger.With("gid", gid)

	// Query group detail
	groupRes, err := r.client.QueryGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err)
		return
	}

	// Check if the user is member in the group
	isMember := groupRes.IsMember(r.context.Config.Granter)
	if !isMember {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	// Get group data
	group, err := r.context.Store.GetGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err)
		return
	}

	// Get all one time public keys in the group
	oneTimePubKeys := make(tss.PublicKeys, groupRes.Group.Size_)
	for _, data := range groupRes.AllRound1Data {
		oneTimePubKeys[data.MemberID-1] = data.OneTimePubKey
	}

	// Compute encrypted secret shares
	encSecretShares, err := tss.ComputeEncryptedSecretShares(
		group.MemberID,
		group.OneTimePrivKey,
		oneTimePubKeys,
		group.Coefficients,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to genrate encrypted secret shares: %s", err)
		return
	}

	// Generate message for round 2
	msg := &types.MsgSubmitDKGRound2{
		GroupID: gid,
		Round2Data: types.Round2Data{
			MemberID:              group.MemberID,
			EncryptedSecretShares: encSecretShares,
		},
		Member: r.context.Config.Granter,
	}

	r.context.MsgCh <- msg
}

// Start starts the Round2 worker.
// It subscribes to events and starts processing incoming events.
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
