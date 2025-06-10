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
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// Round2 is a worker responsible for round2 in the DKG process of tss module
type Round2 struct {
	context *context.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
	reqID   uint64
}

var _ cylinder.Worker = &Round2{}

// NewRound2 creates a new instance of the Round2 worker.
// It initializes the necessary components and returns the created Round2 instance or an error if initialization fails.
func NewRound2(ctx *context.Context) (*Round2, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &Round2{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Round2"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round1_success events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round2) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"tm.event = 'NewBlock' AND %s.%s EXISTS",
		types.EventTypeRound1Success,
		types.AttributeKeyGroupID,
	)
	r.eventCh, err = r.client.Subscribe("Round2", subscriptionQuery, 1000)
	return
}

// handleABCIEvents handles the end block events.
func (r *Round2) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeRound1Success {
			event, err := ParseEvent(sdk.StringEvents{ev}, types.EventTypeRound1Success)
			if err != nil {
				r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			go r.handleGroup(event.GroupID)
		}
	}
}

// handlePendingGroups processes the pending groups.
func (r *Round2) handlePendingGroups() {
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
func (r *Round2) handleGroup(gid tss.GroupID) {
	since := time.Now()

	logger := r.logger.With("gid", gid)

	// Query group detail
	groupRes, err := r.client.QueryGroup(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err)

		metrics.IncProcessRound2FailureCount(uint64(gid))
		return
	}

	if groupRes.Group.Status != types.GROUP_STATUS_ROUND_2 {
		return
	}

	// Check if the user is member in the group
	isMember := groupRes.IsMember(r.context.Config.Granter)
	if !isMember {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming group")

	// Get dkg data of the group
	dkg, err := r.context.Store.GetDKG(gid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err)

		metrics.IncProcessRound2FailureCount(uint64(gid))
		return
	}

	// Get all one time public keys in the group
	oneTimePubKeys := make(tss.Points, groupRes.Group.Size_)
	for _, data := range groupRes.Round1Infos {
		oneTimePubKeys[data.MemberID-1] = data.OneTimePubKey
	}

	// Compute encrypted secret shares
	encSecretShares, err := tss.ComputeEncryptedSecretShares(
		dkg.MemberID,
		dkg.OneTimePrivKey,
		oneTimePubKeys,
		dkg.Coefficients,
		tss.DefaultNonce16Generator{},
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to generate encrypted secret shares: %s", err)

		metrics.IncProcessRound2FailureCount(uint64(gid))
		return
	}

	// Generate message for round 2
	r.reqID += 1
	logger.Info(":delivery_truck: Forward MsgSubmitDKGRound2 to sender with ID: %d", r.reqID)

	r.context.MsgRequestCh <- msg.NewRequest(
		msg.RequestTypeCreateGroupRound2,
		r.reqID,
		types.NewMsgSubmitDKGRound2(
			gid,
			types.Round2Info{
				MemberID:              dkg.MemberID,
				EncryptedSecretShares: encSecretShares,
			},
			r.context.Config.Granter,
		),
		3,
	)

	metrics.ObserveProcessRound2Time(uint64(gid), time.Since(since).Seconds())
	metrics.IncProcessRound2SuccessCount(uint64(gid))
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

	r.handlePendingGroups()

	for ev := range r.eventCh {
		go r.handleABCIEvents(ev.Data.(tmtypes.EventDataNewBlock).ResultFinalizeBlock.Events)
	}
}

// Stop stops the Round2 worker.
func (r *Round2) Stop() error {
	r.logger.Info("stop")
	return r.client.Stop()
}
