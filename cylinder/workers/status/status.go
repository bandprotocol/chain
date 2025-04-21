package status

import (
	"time"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// Status is a worker responsible for checking the status of the active member of the given group and address.
type Status struct {
	context *context.Context
	logger  *logger.Logger
	client  *client.Client
}

var _ cylinder.Worker = &Status{}

// New creates a new instance of the Status worker.
// It initializes the necessary components and returns the created Status instance or an error if initialization fails.
func New(ctx *context.Context) (*Status, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &Status{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Status"),
		client:  cli,
	}, nil
}

// queryStatusWithLogging queries the status of the active member of the given group and address.
// It logs the status of the active member.
func (s *Status) queryStatusWithLogging() {
	address := s.context.Config.Granter

	groupIDs, err := s.context.Store.GetAllGroupIDs()
	if err != nil {
		s.logger.Error(":x: failed to query status: %s", err)
		return
	}

	for _, groupID := range groupIDs {
		members, err := s.client.QueryMembers(groupID)
		if err != nil {
			s.logger.Error(":x: failed to query members: %s", err)
			return
		}

		isActive, err := members.IsActive(address)
		if err != nil {
			s.logger.Error(":x: failed to get member status: %s", err)
			return
		}

		status := ":white_check_mark:"
		if !isActive {
			status = ":x:"
		}
		s.logger.Info("group %d with member %s is active: %s", groupID, s.context.Config.Granter, status)
	}
}

// Start starts the Status worker.
// It queries the status of the active member of the given group and address.
func (s *Status) Start() {
	s.logger.Info("start")

	for {
		s.queryStatusWithLogging()
		time.Sleep(s.context.Config.CheckStatusInterval)
	}
}

// Stop stops the Status worker.
func (s *Status) Stop() error {
	s.logger.Info("stop")
	return s.client.Stop()
}
