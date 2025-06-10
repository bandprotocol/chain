package group

import (
	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/context"
)

// Group is a worker responsible for group creation process of tss module
type Group struct {
	context *context.Context
	workers []cylinder.Worker
}

var _ cylinder.Worker = &Group{}

// New creates a new instance of the Group worker.
// It initializes the necessary components and returns the created Group instance or an error if initialization fails.
func New(ctx *context.Context) (*Group, error) {
	round1, err := NewRound1(ctx)
	if err != nil {
		return nil, err
	}

	round2, err := NewRound2(ctx)
	if err != nil {
		return nil, err
	}

	round3, err := NewRound3(ctx)
	if err != nil {
		return nil, err
	}

	workers := []cylinder.Worker{round1, round2, round3}

	return &Group{
		context: ctx,
		workers: workers,
	}, nil
}

// Start starts the Group worker.
// It start worker of each round of group creation process.
func (g *Group) Start() {
	for _, w := range g.workers {
		go w.Start()
	}
}

// Stop stops the each round's worker.
func (g *Group) Stop() error {
	for _, w := range g.workers {
		if err := w.Stop(); err != nil {
			return err
		}
	}

	return nil
}
