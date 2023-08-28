package replacer

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Replacer is a worker responsible for updating group information of TSS module
type Replacer struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Replacer{}

// New creates a new instance of the Replacer worker.
// It initializes the necessary components and returns the created Replacer instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Replacer, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Replacer{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Replacer"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the replace_success events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Replacer) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"tm.event = 'NewBlock' AND %s.%s EXISTS",
		types.EventTypeReplaceSuccess,
		types.AttributeKeySigningID,
	)
	r.eventCh, err = r.client.Subscribe("Replacer", subscriptionQuery, 1000)
	return
}

// handleABCIEvents handles the end block events.
func (r *Replacer) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeReplaceSuccess {
			event, err := ParseEvent(sdk.StringEvents{ev})
			if err != nil {
				r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			go r.handleReplacement(event.FromGroupID, event.ToGroupID)
		}
	}
}

// handleReplacement processes a replacement.
func (r *Replacer) handleReplacement(fromGroupID tss.GroupID, toGroupID tss.GroupID) {
	// Delete "to" group if any
	if _, err := r.context.Store.GetGroup(toGroupID); err == nil {
		// Delete original group
		r.logger.With("gid", toGroupID).Info(":delivery_truck: Deleting group information")
		r.context.Store.DeleteGroup(toGroupID)
	}

	// Replace "from" group to "to" group if any
	if group, err := r.context.Store.GetGroup(fromGroupID); err == nil {
		r.logger.With("from_gid", fromGroupID).
			With("to_gid", toGroupID).
			Info(":delivery_truck: Replacing group information")
		r.context.Store.SetGroup(toGroupID, group)
	}
}

// Start starts the Replacer worker.
// It subscribes to events and starts processing incoming events.
func (r *Replacer) Start() {
	r.logger.Info("start")

	err := r.subscribe()
	if err != nil {
		r.context.ErrCh <- err
		return
	}

	for ev := range r.eventCh {
		go r.handleABCIEvents(ev.Data.(tmtypes.EventDataNewBlock).ResultEndBlock.Events)
	}
}

// Stop stops the Replacer worker.
func (r *Replacer) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
