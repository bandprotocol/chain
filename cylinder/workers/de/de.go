package de

import (
	"encoding/hex"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// DE is a worker responsible for generating own nonce (DE) of signing process
type DE struct {
	context       *cylinder.Context
	logger        *logger.Logger
	client        *client.Client
	assignEventCh <-chan ctypes.ResultEvent
	useEventCh    <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &DE{}

// New creates a new instance of the DE worker.
// It initializes the necessary components and returns the created DE instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*DE, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
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

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (de *DE) handleTxResult(txResult abci.TxResult) {
	msgLogs, err := event.GetMessageLogs(txResult)
	if err != nil {
		de.logger.Error("Failed to get message logs: %s", err)
		return
	}

	for _, log := range msgLogs {
		events, err := ParseSubmitSignEvents(log.Events)
		if err != nil {
			de.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
			return
		}

		for _, event := range events {
			go de.deleteDE(event.PubDE)
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
	privDEs, err := GenerateDEs(de.context.Config.MinDE, de.context.Config.RandomSecret)
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
			go de.handleTxResult(ev.Data.(tmtypes.EventDataTx).TxResult)
		}
	}()

	// Update if there is assigned DE event.
	for range de.assignEventCh {
		go de.updateDE()
	}
}

// Stop stops the DE worker.
func (de *DE) Stop() error {
	de.logger.Info("stop")
	return de.client.Stop()
}
