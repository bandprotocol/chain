package status

import (
	"fmt"
	"time"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
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

// checkStatus queries the status of the active member from the current and incoming groups.
// It logs the status of the active member.
func (s *Status) checkStatus() {
	address := s.context.Config.Granter

	member, err := s.client.QueryMember(address)
	if err != nil {
		s.logger.Error("failed to query member: %s", err)
		return
	}

	memberResponse := client.NewMemberResponse(member)
	if len(memberResponse.Members) == 0 {
		s.logger.Warn("Member not found in the bandtss current group or incoming group %s", address)
		return
	}

	for _, member := range memberResponse.Members {
		if !member.IsActive {
			s.logger.Warn(":warning:group %d with member %s is inactive", member.GroupID, address)
		} else {
			s.logger.Debug(":white_check_mark:group %d with member %s is active", member.GroupID, address)
		}

		metrics.SetMemberStatus(uint64(member.GroupID), member.IsActive)
	}
}

// Start starts the Status worker.
// It queries the status of the active member of the given group and address.
func (s *Status) Start() {
	s.logger.Info("start")

	ticker := time.NewTicker(s.context.Config.CheckStatusInterval)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("check status")
		s.checkStatus()
	}
}

// Stop stops the Status worker.
func (s *Status) Stop() error {
	s.logger.Info("Stopping status worker")
	return s.client.Stop()
}
