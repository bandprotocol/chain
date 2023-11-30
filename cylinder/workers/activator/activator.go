package activator

import (
	"errors"
	"time"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Activator is a worker responsible for updating active status to the chain
type Activator struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
}

var _ cylinder.Worker = &Activator{}

// New creates a new instance of the Activator worker.
// It initializes the necessary components and returns the created Activator instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Activator, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Activator{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Activator"),
		client:  cli,
	}, nil
}

// updateActive updates last active
func (a *Activator) updateActive() {
	// Query Active information
	status, err := a.client.QueryStatus(a.context.Config.Granter)
	if err != nil {
		a.logger.Error(":cold_sweat: Failed to query status information: %s", err)
		return
	}

	if status.Status != types.MEMBER_STATUS_ACTIVE && status.Status != types.MEMBER_STATUS_PAUSED {
		a.context.ErrCh <- errors.New("The status of the address is not active / paused")
		return
	}

	if time.Now().Before(status.LastActive.Add(a.context.Config.ActivePeriod)) {
		return
	}

	// Log
	a.logger.Info(":delivery_truck: Updating last active")

	// Send MsgActive
	a.context.MsgCh <- &types.MsgHealthCheck{
		Address: a.context.Config.Granter,
	}
}

// Start starts the activator worker.
// It subscribes to events and starts processing incoming events.
func (a *Activator) Start() {
	a.logger.Info("start")

	// Update one time when starting worker first time.
	a.updateActive()

	for range time.Tick(time.Hour * 1) {
		a.updateActive()
	}
}

// Stop stops the Activator worker.
func (a *Activator) Stop() {
	a.logger.Info("stop")
	a.client.Stop()
}
