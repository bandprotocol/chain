package de

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
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
		"tm.event = 'Tx' AND %s.%s = '%s'",
		types.EventTypeRequestSign,
		types.AttributeKeyMember,
		de.context.Config.Granter,
	)
	de.assignEventCh, err = de.client.Subscribe("DE", subscriptionQuery, 1000)

	subscriptionQuery = fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s = '%s'",
		types.EventTypeSubmitSign,
		types.AttributeKeyMember,
		de.context.Config.Granter,
	)
	de.useEventCh, err = de.client.Subscribe("DE", subscriptionQuery, 1000)

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
		event, err := ParseSubmitSignEvent(log)
		if err != nil {
			de.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
			return
		}

		go de.removeDE(event.PubDE)
	}
}

// removeDE removes the specific DE.
func (de *DE) removeDE(pubDE types.DE) {
	// Log
	logger := de.logger.With("D", hex.EncodeToString(pubDE.PubD), "E", hex.EncodeToString(pubDE.PubE))
	logger.Info(":delivery_truck: Removing DE")

	// Remove DE from storage
	err := de.context.Store.RemoveDE(pubDE)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to remove DE: %s", err)
		return
	}
}

// updateDE updates DE if the remaining DE is too low.
func (de *DE) updateDE() {
	// Query DE information
	deRes, err := de.client.QueryDE(de.context.Config.Granter)
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
	privDEs, pubDEs, err := generateDEPairs(de.context.Config.MinDE)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to generate new DE pairs: %s", err)
		return
	}

	// Store all DEs in the store
	for i, privDE := range privDEs {
		err := de.context.Store.SetDE(pubDEs[i], privDE)
		if err != nil {
			de.logger.Error(":cold_sweat: Failed to set new DE in the store: %s", err)
			return
		}
	}

	// Send MsgDE
	de.context.MsgCh <- &types.MsgSubmitDEs{
		DEs:    pubDEs,
		Member: de.context.Config.Granter,
	}
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
func (de *DE) Stop() {
	de.logger.Info("stop")
	de.client.Stop()
}

// generateDEPairs generates n pairs of DE
func generateDEPairs(n uint64) (privDEs []store.DE, pubDEs []types.DE, err error) {
	for i := uint64(1); i <= n; i++ {
		de, err := tss.GenerateKeyPairs(2)
		if err != nil {
			return nil, nil, err
		}

		privDEs = append(privDEs, store.DE{
			PrivD: de[0].PrivKey,
			PrivE: de[1].PrivKey,
		})

		pubDEs = append(pubDEs, types.DE{
			PubD: de[0].PubKey,
			PubE: de[1].PubKey,
		})
	}

	return privDEs, pubDEs, nil
}
