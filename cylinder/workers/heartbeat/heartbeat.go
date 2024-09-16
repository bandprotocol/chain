package heartbeat

import (
	"errors"
	"time"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
)

// Heartbeat is a worker responsible for updating active status to the chain
type Heartbeat struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
}

var _ cylinder.Worker = &Heartbeat{}

// New creates a new instance of the Heartbeat worker.
// It initializes the necessary components and returns the created Heartbeat instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Heartbeat, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Heartbeat{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Heartbeat"),
		client:  cli,
	}, nil
}

// updateHeartbeat updates last active
func (a *Heartbeat) updateHeartbeat() {
	// Query active information
	res, err := a.client.QueryMember(a.context.Config.Granter)
	if err != nil {
		// maybe because not being a member of the current group or the group is not active
		// or the queried node is down.
		a.logger.Debug(":cold_sweat: Failed to query status information: %s", err)
		return
	}

	members := []bandtsstypes.Member{res.CurrentGroupMember, res.IncomingGroupMember}

	for _, member := range members {
		if member.GroupID == 0 {
			continue
		}

		if !member.IsActive {
			a.context.ErrCh <- errors.New("the status of the address is not active")
			continue
		}

		if time.Now().Before(member.LastActive.Add(a.context.Config.ActivePeriod)) {
			continue
		}

		// Log
		a.logger.Info(":delivery_truck: Updating last active")

		// Send MsgActive
		a.context.MsgCh <- bandtsstypes.NewMsgHeartbeat(a.context.Config.Granter, member.GroupID)
	}
}

// Start starts the heartbeat worker that will check latest heartbeat of validator on BandChain
// and send heartbeat msg if needed every hour.
func (a *Heartbeat) Start() {
	a.logger.Info("start")

	// Update one time when starting worker first time.
	a.updateHeartbeat()

	for range time.Tick(time.Hour * 1) {
		a.updateHeartbeat()
	}
}

// Stop stops the Heartbeat worker.
func (a *Heartbeat) Stop() error {
	a.logger.Info("stop")
	return a.client.Stop()
}
