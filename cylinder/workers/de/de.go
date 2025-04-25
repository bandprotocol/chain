package de

import (
	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/context"
)

var _ cylinder.Worker = &DE{}

// DE is a worker responsible for managing own nonce (DE) that being used for signing process
type DE struct {
	context *context.Context
	workers []cylinder.Worker
}

// New creates a new instance of the DE-related workers.
// It initializes the necessary components and returns the created DE instance or an error if initialization fails.
func New(ctx *context.Context) (*DE, error) {
	return &DE{
		context: ctx,
	}, nil
}

// Start starts the DE-related workers.
func (de *DE) Start() {
	updateDE, err := NewUpdateDE(de.context)
	if err != nil {
		de.context.ErrCh <- err
		return
	}

	deleteDE, err := NewDeleteDE(de.context)
	if err != nil {
		de.context.ErrCh <- err
		return
	}

	de.workers = []cylinder.Worker{updateDE, deleteDE}
	for _, w := range de.workers {
		go w.Start()
	}
}

// Stop stops DE-related workers.
func (de *DE) Stop() error {
	for _, w := range de.workers {
		if err := w.Stop(); err != nil {
			return err
		}
	}

	return nil
}
