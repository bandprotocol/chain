package active

import (
	"errors"
	"time"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Active is a worker responsible for generating own nonce (DE) of signing process
type Active struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
}

var _ cylinder.Worker = &Active{}

// New creates a new instance of the Active worker.
// It initializes the necessary components and returns the created Active instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Active, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Active{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Active"),
		client:  cli,
	}, nil
}

// updateActive updates last active
func (a *Active) updateActive() {
	// Query Active information
	status, err := a.client.QueryStatus(a.context.Config.Granter)
	if err != nil {
		a.logger.Error(":cold_sweat: Failed to query status information: %s", err)
		return
	}

	if status.Status != types.MEMBER_STATUS_ACTIVE {
		a.context.ErrCh <- errors.New("The status of the address is not active")
		return
	}

	if time.Now().Before(status.LastActive.Add(a.context.Config.ActivePeriod)) {
		return
	}

	// Log
	a.logger.Info(":delivery_truck: Updating last active")

	// Send MsgActive
	a.context.MsgCh <- &types.MsgActive{
		Address: a.context.Config.Granter,
	}
}

// Start starts the active worker.
// It subscribes to events and starts processing incoming events.
func (a *Active) Start() {
	a.logger.Info("start")

	// Update one time when starting worker first time.
	a.updateActive()

	for range time.Tick(time.Hour * 1) {
		a.updateActive()
	}
}

// Stop stops the Active worker.
func (a *Active) Stop() {
	a.logger.Info("stop")
	a.client.Stop()
}
