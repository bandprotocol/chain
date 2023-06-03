package de

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// DE is a worker responsible for generating own nonce (DE) of signing process
type DE struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	eventCh <-chan ctypes.ResultEvent
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

// subscribe subscribes to the DE usage events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (de *DE) subscribe() error {
	var err error
	de.eventCh, err = de.client.Subscribe(
		"DE",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s = '%s'",
			types.EventTypeRequestSign,
			types.AttributeKeyMember,
			de.context.Config.Granter,
		),
		1000,
	)
	return err
}

// updateDE updates DE if the remaining of DE is too low.
func (de *DE) updateDE() {
	// Query DE information
	deRes, err := de.client.QueryDE(de.context.Config.Granter)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to query DE information: %s", err.Error())
		return
	}

	// Check remaining of DE, ignore if it's more than min-DE
	remaining := deRes.GetRemaining()
	if remaining >= de.context.Config.MinDE {
		return
	}

	// Log
	de.logger.Info(":delivery_truck: Updating DE")

	// Generate new DEs
	privDEs, pubDEs, err := generateDEPairs(de.context.Config.MinDE)
	if err != nil {
		de.logger.Error(":cold_sweat: Failed to generate new DE: %s", err.Error())
		return
	}

	// Stores all DEs to store
	for i, privDE := range privDEs {
		err := de.context.Store.SetDE(pubDEs[i], privDE)
		if err != nil {
			de.logger.Error(":cold_sweat: Failed to set new DE to store: %s", err.Error())
			return
		}
	}

	// Send MsgDE
	de.context.MsgCh <- &types.MsgSubmitDEPairs{
		DEPairs: pubDEs,
		Member:  de.context.Config.Granter,
	}
}

// Start starts the DE worker.
// It subscribes to DE usage events and starts processing incoming events.
func (de *DE) Start() {
	de.logger.Info("start")

	err := de.subscribe()
	if err != nil {
		de.context.ErrCh <- err
		return
	}

	// Update one time when starting worker first time.
	go de.updateDE()

	for range de.eventCh {
		go de.updateDE()
	}
}

// Stop stops the DE worker.
func (de *DE) Stop() {
	de.logger.Info("stop")
	de.client.Stop()
}

// generateDEPairs generates n pairs of DE
func generateDEPairs(n uint64) ([]store.DE, []types.DE, error) {
	var privDEs []store.DE
	var pubDEs []types.DE

	for i := uint64(1); i <= n; i++ {
		d, err := tss.GenerateKeyPair()
		if err != nil {
			return nil, nil, err
		}

		e, err := tss.GenerateKeyPair()
		if err != nil {
			return nil, nil, err
		}

		privDEs = append(privDEs, store.DE{
			PrivD: d.PrivateKey,
			PrivE: e.PrivateKey,
		})

		pubDEs = append(pubDEs, types.DE{
			PubD: d.PublicKey,
			PubE: e.PublicKey,
		})
	}

	return privDEs, pubDEs, nil
}
