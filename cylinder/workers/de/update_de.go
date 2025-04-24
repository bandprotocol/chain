package de

import (
	"fmt"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// UpdateDE is a worker responsible for updating DEs in the store and chains
type UpdateDE struct {
	context *context.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
	cntUsed uint64
}

var _ cylinder.Worker = &UpdateDE{}

// NewUpdateDE creates a new UpdateDE worker.
func NewUpdateDE(ctx *context.Context) (*UpdateDE, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &UpdateDE{
		context: ctx,
		logger:  ctx.Logger.With("worker", "UpdateDE"),
		client:  cli,
	}, nil
}

// Start starts the UpdateDE worker.
func (u *UpdateDE) Start() {
	u.logger.Info("start")

	if err := u.subscribe(); err != nil {
		u.context.ErrCh <- err
		return
	}

	// Update one time when starting worker first time.
	if err := u.intervalUpdateDE(); err != nil {
		u.context.ErrCh <- err
		return
	}

	// Update DE if there is assigned DE event or DE is used.
	ticker := time.NewTicker(u.context.Config.CheckDEInterval)
	for {
		select {
		case <-ticker.C:
			if err := u.intervalUpdateDE(); err != nil {
				u.logger.Error(":cold_sweat: Failed to do an interval update DE: %s", err)
			}
		case <-u.eventCh:
			u.cntUsed += 1
			if u.cntUsed >= u.context.Config.MinDE {
				u.updateDE(u.cntUsed)
				u.cntUsed = 0
			}
		}
	}
}

// Stop stops the UpdateDE worker.
func (u *UpdateDE) Stop() error {
	u.logger.Info("stop")
	return u.client.Stop()
}

// subscribe subscribes to the events that trigger the DE update.
func (u *UpdateDE) subscribe() (err error) {
	assignedDEQuery := fmt.Sprintf(
		"%s.%s = '%s'",
		types.EventTypeRequestSignature,
		types.AttributeKeyAddress,
		u.context.Config.Granter,
	)

	u.eventCh, err = u.client.Subscribe("AssignedDE", assignedDEQuery, 1000)
	return err
}

// updateDE updates DE if the remaining DE is too low.
func (u *UpdateDE) updateDE(numNewDE uint64) {
	canUpdate, err := u.canUpdateDE()
	if err != nil {
		u.logger.Error(":cold_sweat: Cannot update DE: %s", err)
		return
	}
	if !canUpdate {
		u.logger.Debug(
			":cold_sweat: Cannot update DE: the granter is not a member of the current or incoming group and gas price isn't set in the config",
		)
		return
	}

	u.logger.Info(":delivery_truck: Updating DE")

	// Generate new DE pairs
	privDEs, err := GenerateDEs(
		numNewDE,
		u.context.Config.RandomSecret,
		u.context.Store,
	)
	if err != nil {
		u.logger.Error(":cold_sweat: Failed to generate new DE pairs: %s", err)
		return
	}

	// Store all DEs in the store
	var pubDEs []types.DE
	for _, privDE := range privDEs {
		pubDEs = append(pubDEs, privDE.PubDE)

		if err := u.context.Store.SetDE(privDE); err != nil {
			u.logger.Error(":cold_sweat: Failed to set new DE in the store: %s", err)
			return
		}

		metrics.IncOffChainDELeftGauge()
	}

	u.logger.Info(":white_check_mark: Successfully generated %d new DE pairs", numNewDE)

	// Send MsgDE
	u.context.MsgCh <- types.NewMsgSubmitDEs(pubDEs, u.context.Config.Granter)
}

// canUpdateDE checks if the system allows to update DEs into the system and chain.
func (u *UpdateDE) canUpdateDE() (bool, error) {
	gasPrices, err := sdk.ParseDecCoins(u.context.Config.GasPrices)
	if err != nil {
		u.logger.Debug(":cold_sweat: Failed to parse gas prices from config: %s", err)
	}

	// If the gas price is non-zero, it indicates that the user is willing to pay
	// a transaction fee for submitting DEs to the chain.
	if gasPrices != nil && !gasPrices.IsZero() {
		return true, nil
	}

	// If the address is a member of the current group, the system can submit DEs to the chain
	// without paying gas.
	resp, err := u.client.QueryMember(u.context.Config.Granter)
	if err != nil {
		return false, fmt.Errorf("failed to query member information: %w", err)
	}

	if resp.CurrentGroupMember.Address == u.context.Config.Granter ||
		resp.IncomingGroupMember.Address == u.context.Config.Granter {
		return true, nil
	}

	return false, nil
}

// intervalUpdateDE updates DE on the chain so that the remaining DE is
// always above the minimum threshold.
func (u *UpdateDE) intervalUpdateDE() error {
	deCount, err := u.getDECount()
	if err != nil {
		return err
	}

	if deCount < 2*u.context.Config.MinDE {
		u.updateDE(2*u.context.Config.MinDE - deCount)
		u.cntUsed = 0
	}

	metrics.SetOnChainDELeftGauge(float64(deCount))

	return nil
}

// getDECount queries the number of DEs on the chain.
func (u *UpdateDE) getDECount() (uint64, error) {
	// Query DE information
	deRes, err := u.client.QueryDE(u.context.Config.Granter, 0, 1)
	if err != nil {
		u.logger.Error(":cold_sweat: Failed to query DE information: %s", err)
		return 0, err
	}

	return deRes.GetRemaining(), nil
}
