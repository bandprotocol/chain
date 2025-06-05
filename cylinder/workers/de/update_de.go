package de

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
	"github.com/bandprotocol/chain/v3/cylinder/parser"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// UpdateDE is a worker responsible for updating DEs in the store and chains
type UpdateDE struct {
	context          *context.Context
	logger           *logger.Logger
	client           *client.Client
	eventCh          <-chan ctypes.ResultEvent
	cntUsed          int64
	maxDESizeOnChain uint64
}

var _ cylinder.Worker = &UpdateDE{}

// NewUpdateDE creates a new UpdateDE worker.
func NewUpdateDE(ctx *context.Context) (*UpdateDE, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	params, err := cli.QueryTssParams()
	if err != nil {
		return nil, err
	}

	return &UpdateDE{
		context:          ctx,
		logger:           ctx.Logger.With("worker", "UpdateDE"),
		client:           cli,
		maxDESizeOnChain: params.MaxDESize,
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
		case resultEvent := <-u.eventCh:
			deUsed := int64(0)
			var err error
			switch data := resultEvent.Data.(type) {
			case tmtypes.EventDataTx:
				deUsed, err = u.countCreatedSignings(data.Result.Events)
			case tmtypes.EventDataNewBlock:
				deUsed, err = u.countCreatedSignings(data.ResultFinalizeBlock.Events)
			default:
				continue
			}

			if err != nil {
				u.logger.Error(":cold_sweat: Failed to count created signings: %s", err)
				continue
			}
			u.cntUsed += deUsed

			// if the system used DE over the threshold, add new DEs
			// the threshold plus the expectedDESize (2/3 maxDESizeOnChain) shouldn't be over
			// the maxDESizeOnChain to prevent any transaction revert.
			// The maxDESize params should be set to be at least 3 times of the normal usage (per block).
			threshold := u.maxDESizeOnChain / 3
			if u.cntUsed >= int64(threshold) {
				u.logger.Info(":delivery_truck: DEs are used over the threshold, adding new DEs")
				if err := u.updateDE(threshold); err != nil {
					u.logger.Error(":cold_sweat: Failed to update DE: %s", err)
					continue
				}

				u.cntUsed -= int64(threshold)
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
func (u *UpdateDE) updateDE(numNewDE uint64) error {
	isTssMember, err := u.isTssMember()
	if err != nil {
		return fmt.Errorf("isTssMember error: %s", err)
	}

	isGasPriceSet, err := u.isGasPriceSet()
	if err != nil {
		return fmt.Errorf("isGasPriceSet error: %s", err)
	}

	if !isTssMember && !isGasPriceSet {
		u.logger.Debug(":cold_sweat: Skip updating DE; not a tss member and gas price isn't set")
		return nil
	}

	u.logger.Info(":delivery_truck: Updating DE")

	// Generate new DE pairs
	privDEs, err := GenerateDEs(
		numNewDE,
		u.context.Config.RandomSecret,
		u.context.Store,
	)
	if err != nil {
		return fmt.Errorf("failed to generate new DE pairs: %s", err)
	}

	// Store all DEs in the store
	var pubDEs []types.DE
	for _, privDE := range privDEs {
		pubDEs = append(pubDEs, privDE.PubDE)

		if err := u.context.Store.SetDE(privDE); err != nil {
			return fmt.Errorf("failed to set new DE in the store: %s", err)
		}

		metrics.IncOffChainDELeftGauge()
	}

	u.logger.Info(":white_check_mark: Successfully generated %d new DE pairs", numNewDE)

	// Send MsgDE
	u.context.MsgCh <- types.NewMsgSubmitDEs(pubDEs, u.context.Config.Granter)
	return nil
}

// isTssMember checks if the granter is a tss member.
func (u *UpdateDE) isTssMember() (bool, error) {
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

// isGasPriceSet checks if the gas price is set.
func (u *UpdateDE) isGasPriceSet() (bool, error) {
	gasPrices, err := sdk.ParseDecCoins(u.context.Config.GasPrices)
	if err != nil {
		return false, fmt.Errorf("failed to parse gas prices from config: %w", err)
	}

	// If the gas price is non-zero, it indicates that the user is willing to pay
	// a transaction fee for submitting DEs to the chain.
	if gasPrices != nil && !gasPrices.IsZero() {
		return true, nil
	}

	return false, nil
}

// intervalUpdateDE updates DE on the chain so that the remaining DE is
// always above the minimum threshold.
func (u *UpdateDE) intervalUpdateDE() error {
	// also update the maxDESizeOnChain
	params, err := u.client.QueryTssParams()
	if err != nil {
		return err
	}
	u.maxDESizeOnChain = params.MaxDESize

	deCount, err := u.getDECount()
	if err != nil {
		return err
	}

	expectedDESizeOnChain := (2 * u.maxDESizeOnChain) / 3
	if deCount < expectedDESizeOnChain {
		if err := u.updateDE(expectedDESizeOnChain - deCount); err != nil {
			return err
		}

		u.cntUsed = 0
	} else {
		// can be negative to represent that the system has more DEs than expected
		u.cntUsed = int64(expectedDESizeOnChain) - int64(deCount)
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

// countCreatedSignings counts the number of signings created from the given events.
func (u *UpdateDE) countCreatedSignings(abciEvents []abci.Event) (int64, error) {
	cnt := int64(0)
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type != types.EventTypeRequestSignature {
			continue
		}

		signatureEvents, err := parser.ParseRequestSignatureEvents(sdk.StringEvents{ev})
		if err != nil {
			return 0, err
		}

		for _, signatureEvent := range signatureEvents {
			ok, err := u.isGranterSigner(signatureEvent.SigningID)
			if err != nil {
				u.logger.Error(
					":cold_sweat: isGranterSigner Failed at SigningID %d: %s",
					signatureEvent.SigningID,
					err,
				)
				continue
			}
			if ok {
				cnt += 1
			}
		}
	}

	return cnt, nil
}

// isGranterSigner checks if the granter is the assigned member of the signing.
func (u *UpdateDE) isGranterSigner(signingID tss.SigningID) (bool, error) {
	signingRes, err := u.client.QuerySigning(signingID)
	if err != nil {
		return false, err
	}

	if _, err := signingRes.GetAssignedMember(u.context.Config.Granter); err != nil {
		return false, nil
	}

	return true, nil
}
