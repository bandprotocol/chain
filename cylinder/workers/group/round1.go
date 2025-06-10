package group

import (
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/cylinder/msg"
	"github.com/bandprotocol/chain/v3/cylinder/store"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// Round1 is a worker responsible for round1 in the DKG process of tss module
type Round1 struct {
	context *context.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
	reqID   uint64
}

var _ cylinder.Worker = &Round1{}

// NewRound1 creates a new instance of the Round1 worker.
// It initializes the necessary components and returns the created Round1 instance or an error if initialization fails.
func NewRound1(ctx *context.Context) (*Round1, error) {
	// create http client
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &Round1{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Round1"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the create_group events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round1) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"tm.event = 'NewBlock' AND %s.%s = '%s'",
		types.EventTypeCreateGroup,
		types.AttributeKeyAddress,
		r.context.Config.Granter,
	)
	r.eventCh, err = r.client.Subscribe("Round1", subscriptionQuery, 1000)
	return
}

// handleABCIEvents handles the end block events.
func (r *Round1) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeCreateGroup {
			event, err := ParseEvent(sdk.StringEvents{ev}, types.EventTypeCreateGroup)
			if err != nil {
				r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			go r.handleGroup(event.GroupID)
		}
	}
}

// handlePendingGroups processes the pending groups.
func (r *Round1) handlePendingGroups() {
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
func (r *Round1) handleGroup(gid tss.GroupID) {
	since := time.Now()

	logger := r.logger.With("gid", gid)

	// Query group detail
	groupRes, err := r.client.QueryGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err)

		metrics.IncProcessRound1FailureCount(uint64(gid))
		return
	}

	if groupRes.Group.Status != types.GROUP_STATUS_ROUND_1 {
		return
	}

	// Check if the user is member in the group
	mid, err := groupRes.GetMemberID(r.context.Config.Granter)
	if err != nil {
		metrics.IncProcessRound1FailureCount(uint64(gid))
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	// Generate round1 data
	data, err := tss.GenerateRound1Info(mid, groupRes.Group.Threshold, groupRes.DKGContext)
	if err != nil {
		logger.Error(":cold_sweat: Failed to generate round1 data with error: %s", err)

		metrics.IncProcessRound1FailureCount(uint64(gid))
		return
	}

	// Set group data
	dkg := store.DKG{
		GroupID:        gid,
		MemberID:       mid,
		Coefficients:   data.Coefficients,
		OneTimePrivKey: data.OneTimePrivKey,
	}
	err = r.context.Store.SetDKG(dkg)
	if err != nil {
		logger.Error(":cold_sweat: Failed to set DKG with error: %s", err)

		metrics.IncProcessRound1FailureCount(uint64(gid))
		return
	}

	metrics.IncDKGLeftGauge()

	// Send the message to the message channel
	r.reqID += 1
	logger.Info(":delivery_truck: Forward MsgSubmitDKGRound1 to sender with ID: %d", r.reqID)

	r.context.MsgRequestCh <- msg.NewRequest(
		msg.RequestTypeCreateGroupRound1,
		r.reqID,
		types.NewMsgSubmitDKGRound1(
			gid,
			types.Round1Info{
				MemberID:           mid,
				CoefficientCommits: data.CoefficientCommits,
				OneTimePubKey:      data.OneTimePubKey,
				A0Signature:        data.A0Signature,
				OneTimeSignature:   data.OneTimeSignature,
			},
			r.context.Config.Granter,
		),
		3,
	)

	metrics.ObserveProcessRound1Time(uint64(gid), time.Since(since).Seconds())
	metrics.IncProcessRound1SuccessCount(uint64(gid))
}

// Start starts the Round1 worker.
// It subscribes to the events, and continuously processes incoming events by calling handleABCIEvents.
func (r *Round1) Start() {
	r.logger.Info("start")

	err := r.subscribe()
	if err != nil {
		r.context.ErrCh <- err
		return
	}

	r.handlePendingGroups()

	for ev := range r.eventCh {
		go r.handleABCIEvents(ev.Data.(tmtypes.EventDataNewBlock).ResultFinalizeBlock.Events)
	}
}

// Stop stops the Round1 worker.
func (r *Round1) Stop() error {
	r.logger.Info("stop")
	return r.client.Stop()
}
