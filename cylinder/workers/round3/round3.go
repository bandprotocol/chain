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

	logger *logger.Logger
	client *client.Client

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
		logger:  ctx.Logger.With("worker", "round3"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round2 events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round3) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round3",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS",
			types.EventTypeRound2Success,
			types.AttributeKeyGroupID,
		),
		1000,
	)
	return err
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round3) handleTxResult(txResult abci.TxResult) {
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

		go r.handleGroup(event.GroupID)
	}
}

// handleEvent processes an incoming group.
func (r *Round3) handleGroup(gid tss.GroupID) {
	logger := r.logger.With("gid", gid)

	// Query group detail
	groupRes, err := r.client.QueryGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err.Error())
		return
	}

	// Check if the user is member in the group
	isMember := groupRes.IsMember(r.context.Config.Granter)
	if !isMember {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	// Set group data
	group, err := r.context.Store.GetGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err.Error())
		return
	}

	// Get own private key
	ownPrivKey, complains, err := getOwnPrivKey(group, groupRes)
	if err != nil {
		logger.Error(":cold_sweat: Failed to get own private key or complains: %s", err.Error())
		return
	}

	// If there is any complain, send MsgComplain
	if len(complains) > 0 {
		// Send message complains
		r.context.MsgCh <- &types.MsgComplain{
			GroupID:   gid,
			Complains: complains,
			Member:    r.context.Config.Granter,
		}
		return
	}

	// Generate own private key and update it in store
	group.PrivKey = ownPrivKey
	r.context.Store.SetGroup(gid, group)
	ownPubKey, err := ownPrivKey.PublicKey()
	if err != nil {
		logger.Error(":cold_sweat: Failed to generate own public key: %s", err.Error())
		return
	}

	// Sign own public key
	ownPubKeySig, err := tss.SignOwnPublickey(
		group.MemberID,
		groupRes.DKGContext,
		ownPubKey,
		ownPrivKey,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to sign own public key: %s", err.Error())
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
// It subscribes to round2 events and starts processing incoming events.
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
