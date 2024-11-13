package de

import (
	"encoding/hex"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

var _ cylinder.Worker = &DE{}

// DE is a worker responsible for generating own nonce (DE) of signing process
type DE struct {
	context       *context.Context
	logger        *logger.Logger
	client        *client.Client
	assignEventCh <-chan ctypes.ResultEvent
	useEventCh    <-chan ctypes.ResultEvent
}

// New creates a new instance of the DE worker.
// It initializes the necessary components and returns the created DE instance or an error if initialization fails.
func New(ctx *context.Context) (*DE, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &DE{
		context: ctx,
		logger:  ctx.Logger.With("worker", "DE"),
		client:  cli,
	}, nil
}

// subscribe subscribes to request_sign events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (de *DE) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"%s.%s = '%s'",
		types.EventTypeRequestSignature,
		types.AttributeKeyAddress,
		de.context.Config.Granter,
	)
	de.assignEventCh, err = de.client.Subscribe("DE-assigned", subscriptionQuery, 1000)
	if err != nil {
		return
	}

	subscriptionQuery = fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s = '%s'",
		types.EventTypeSubmitSignature,
		types.AttributeKeyAddress,
		de.context.Config.Granter,
	)
	de.useEventCh, err = de.client.Subscribe("DE-submitted", subscriptionQuery, 1000)

	return
}

// handleABCIEvents signs the specific signingID if the given events contain a request_signature event.
func (de *DE) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeSubmitSignature {
			events, err := ParseSubmitSignEvents(sdk.StringEvents{ev})
			if err != nil {
				de.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			for _, event := range events {
				go de.deleteDE(event.PubDE)
			}
		}
	}
}

// deleteDE deletes the specific DE.
func (de *DE) deleteDE(pubDE types.DE) {
	// Log
	logger := de.logger.With("D", hex.EncodeToString(pubDE.PubD), "E", hex.EncodeToString(pubDE.PubE))
	logger.Info(":delivery_truck: Removing DE")

	// Remove DE from storage
	err := de.context.Store.DeleteDE(pubDE)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to remove DE: %s", err)
		return
	}
}

// updateDE updates DE if the remaining DE is too low.
func (de *DE) updateDE() {
	// Query DE information
	deRes, err := de.client.QueryDE(de.context.Config.Granter, 0, 1)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to query DE information: %s", err)
		return
	}

	// Check remaining DE, ignore if it's more than min-DE
	remaining := deRes.GetRemaining()
	if remaining >= de.context.Config.MinDE {
		return
	}

	// Log
	de.logger.Info(":delivery_truck: Updating DE")

	// Generate new DE pairs
	privDEs, err := GenerateDEs(
		de.context.Config.MinDE,
		de.context.Config.RandomSecret,
		de.context.Store,
	)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to generate new DE pairs: %s", err)
		return
	}

	// Store all DEs in the store
	var pubDEs []types.DE
	for _, privDE := range privDEs {
		pubDEs = append(pubDEs, privDE.PubDE)

		if err := de.context.Store.SetDE(privDE); err != nil {
			de.logger.Error(":cold_sweat: Failed to set new DE in the store: %s", err)
			return
		}
	}

	// Send MsgDE
	de.context.MsgCh <- types.NewMsgSubmitDEs(pubDEs, de.context.Config.Granter)
}

// Start starts the DE worker.
// It subscribes to events and starts processing incoming events.
func (de *DE) Start() {
	de.logger.Info("start")

	err := de.subscribe()
	if err != nil {
		de.context.ErrCh <- err
		return
	}

	// Update one time when starting worker first time.
	de.updateDE()

	// Remove DE if there is used DE event.
	go func() {
		for ev := range de.useEventCh {
			go de.handleABCIEvents(ev.Data.(tmtypes.EventDataTx).TxResult.Result.Events)
		}
	}()

	// Update DE if there is assigned DE event or DE is used.
	ticker := time.NewTicker(de.context.Config.CheckingDEInterval)
	for {
		select {
		case <-ticker.C:
			de.updateDE()
		case <-de.assignEventCh:
			de.updateDE()
		}
	}
}

// Stop stops the DE worker.
func (de *DE) Stop() error {
	de.logger.Info("stop")
	return de.client.Stop()
}
