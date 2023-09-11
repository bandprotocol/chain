package group

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Round3 is a worker responsible for round3 in the DKG process of TSS module
type Round3 struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round3{}

// NewRound3 creates a new instance of the Round3 worker.
// It initializes the necessary components and returns the created Round3 instance or an error if initialization fails.
func NewRound3(ctx *cylinder.Context) (*Round3, error) {
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
		"tm.event = 'NewBlock' AND %s.%s EXISTS",
		types.EventTypeRound2Success,
		types.AttributeKeyGroupID,
	)
	r.eventCh, err = r.client.Subscribe("Round3", subscriptionQuery, 1000)
	return
}

// handleABCIEvents handles the end block events.
func (r *Round3) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeRound2Success {
			event, err := ParseEvent(sdk.StringEvents{ev}, types.EventTypeRound2Success)
			if err != nil {
				r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			go r.handleGroup(event.GroupID)
		}
	}
}

// handlePendingGroups processes the pending groups.
func (r *Round3) handlePendingGroups() {
	res, err := r.client.QueryPendingGroups(r.context.Config.Granter)
	if err != nil {
		r.logger.Error(":cold_sweat: Failed to get pending groups: %s", err)
		return
	}

	for _, gid := range res.PendingGroups {
		go r.handleGroup(tss.GroupID(gid))
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

	if groupRes.Group.Status != types.GROUP_STATUS_ROUND_3 {
		return
	}

	// Check if the user is member in the group
	if !groupRes.IsMember(r.context.Config.Granter) {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	group, err := r.context.Store.GetGroup(groupRes.Group.PubKey)
	if err != nil {
		// Set DKG data
		dkg, err := r.context.Store.GetDKG(gid)
		if err != nil {
			logger.Error(":cold_sweat: Failed to find group in store: %s", err)
			return
		}

		// Get own private key
		ownPrivKey, complaints, err := getOwnPrivKey(dkg, groupRes)
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
				Address:    r.context.Config.Granter,
			}
			return
		}

		// Generate own private key and update it in store
		group = store.Group{
			MemberID: dkg.MemberID,
			PrivKey:  ownPrivKey,
		}

		err = r.context.Store.SetGroup(groupRes.Group.PubKey, group)
		if err != nil {
			logger.Error(":cold_sweat: Failed to set group with error: %s", err)
			return
		}

		err = r.context.Store.DeleteDKG(gid)
		if err != nil {
			logger.Error(":cold_sweat: Failed to delete DKG with error: %s", err)
			return
		}
	}

	// Sign own public key
	ownPubKeySig, err := tss.SignOwnPubkey(
		group.MemberID,
		groupRes.DKGContext,
		group.PrivKey.Point(),
		group.PrivKey,
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
		Address:      r.context.Config.Granter,
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

	r.handlePendingGroups()

	for ev := range r.eventCh {
		go r.handleABCIEvents(ev.Data.(tmtypes.EventDataNewBlock).ResultEndBlock.Events)
	}
}

// Stop stops the Round3 worker.
func (r *Round3) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
