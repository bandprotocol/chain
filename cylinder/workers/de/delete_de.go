package de

import (
	"encoding/hex"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/cylinder/msg"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// DeleteDE is a worker responsible for deleting DEs from the store once being triggered by the related chain events
type DeleteDE struct {
	context                *context.Context
	logger                 *logger.Logger
	client                 *client.Client
	deleteDEEventCh        <-chan ctypes.ResultEvent
	submitSignatureEventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &DeleteDE{}

// NewDeleteDE creates a new DeleteDE worker.
func NewDeleteDE(ctx *context.Context) (*DeleteDE, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	return &DeleteDE{
		context: ctx,
		logger:  ctx.Logger.With("worker", "DeleteDE"),
		client:  cli,
	}, nil
}

// Start starts the DeleteDE worker.
func (d *DeleteDE) Start() {
	d.logger.Info("start")

	if err := d.subscribe(); err != nil {
		d.context.ErrCh <- err
		return
	}

	for {
		select {
		case ev := <-d.deleteDEEventCh:
			go d.deleteDEFromABCIEvents(ev.Data.(tmtypes.EventDataTx).Result.Events)
		case ev := <-d.submitSignatureEventCh:
			go d.deleteDEFromABCIEvents(ev.Data.(tmtypes.EventDataTx).Result.Events)
		}
	}
}

// Stop stops the DeleteDE worker.
func (d *DeleteDE) Stop() error {
	d.logger.Info("stop")
	return d.client.Stop()
}

// subscribe subscribes to the events that trigger the DEs deletion.
func (d *DeleteDE) subscribe() (err error) {
	deleteDEQuery := fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s = '%s'",
		types.EventTypeDEDeleted,
		types.AttributeKeyAddress,
		d.context.Config.Granter,
	)

	d.deleteDEEventCh, err = d.client.Subscribe("DeleteDE", deleteDEQuery, 1000)
	if err != nil {
		return err
	}

	submitSignatureQuery := fmt.Sprintf(
		"tm.event = 'Tx' AND %s.%s = '%s'",
		types.EventTypeSubmitSignature,
		types.AttributeKeyAddress,
		d.context.Config.Granter,
	)

	d.submitSignatureEventCh, err = d.client.Subscribe("SubmitSignature", submitSignatureQuery, 1000)
	if err != nil {
		return err
	}

	return nil
}

// deleteDEFromABCIEvents signs the specific signingID if the given events contain a request_signature event.
func (d *DeleteDE) deleteDEFromABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeSubmitSignature || ev.Type == types.EventTypeDEDeleted {
			pubDEs, err := ParsePubDEFromEvents(sdk.StringEvents{ev}, ev.Type)
			if err != nil {
				d.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			for _, pubDE := range pubDEs {
				go d.deleteDE(pubDE.PubDE)
			}
		}
	}
}

// deleteDE deletes the specific DE.
func (d *DeleteDE) deleteDE(pubDE types.DE) {
	// Log
	logger := d.logger.With("D", hex.EncodeToString(pubDE.PubD), "E", hex.EncodeToString(pubDE.PubE))
	logger.Info(":delivery_truck: Removing DE")

	// Remove DE from storage
	err := d.context.Store.DeleteDE(pubDE)
	if err != nil {
		d.logger.Error(":cold_sweat: Failed to remove DE: %s", err)
		return
	}

	metrics.DecOffChainDELeftGauge()
}

// GetResponseReceivers returns the message response receivers of the worker.
func (d *DeleteDE) GetResponseReceivers() []*msg.ResponseReceiver {
	return nil
}
