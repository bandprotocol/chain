package round3

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

// Round3 is a worker responsible for round3 in the DKG process of TSS module
type Round3 struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round3{}

// New creates a new instance of the Round3 worker.
// It initializes the necessary components and returns the created Round3 instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Round3, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round3{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Round3"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round2_success events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round3) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s EXISTS",
		types.EventTypeRound2Success,
		types.AttributeKeyGroupID,
	)
	r.eventCh, err = r.client.Subscribe("Round3", subscriptionQuery, 1000)
	return
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round3) handleTxResult(txResult abci.TxResult) {
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
func (r *Round3) handleGroup(gid tss.GroupID) {
	logger := r.logger.With("gid", gid)

	// Query group detail
	groupRes, err := r.client.QueryGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err)
		return
	}

	// Check if the user is member in the group
	if !groupRes.IsMember(r.context.Config.Granter) {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	// Set group data
	group, err := r.context.Store.GetGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err)
		return
	}

	// Get own private key
	ownPrivKey, complaints, err := getOwnPrivKey(group, groupRes)
	if err != nil {
		logger.Error(":cold_sweat: Failed to get own private key or complaints: %s", err)
		return
	}

	// If there is any complaint, send MsgComplain
	if len(complaints) > 0 {
		// Send message complaints
		r.context.MsgCh <- &types.MsgComplain{
			GroupID:    gid,
			Complaints: complaints,
			Member:     r.context.Config.Granter,
		}
		return
	}

	// Generate own private key and update it in store
	group.PrivKey = ownPrivKey
	group.PubKey = groupRes.Group.PubKey
	r.context.Store.SetGroup(gid, group)

	// Get own public key
	ownPubKey, err := ownPrivKey.PublicKey()
	if err != nil {
		logger.Error(":cold_sweat: Failed to get own public key: %s", err)
		return
	}

	// Sign own public key
	ownPubKeySig, err := tss.SignOwnPubkey(
		group.MemberID,
		groupRes.DKGContext,
		ownPubKey,
		ownPrivKey,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to sign own public key: %s", err)
		return
	}

	// Send MsgConfirm
	r.context.MsgCh <- &types.MsgConfirm{
		GroupID:      gid,
		MemberID:     group.MemberID,
		OwnPubKeySig: ownPubKeySig,
		Member:       r.context.Config.Granter,
	}
}

// Start starts the Round3 worker.
// It subscribes to events and starts processing incoming events.
func (r *Round3) Start() {
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

// Stop stops the Round3 worker.
func (r *Round3) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
