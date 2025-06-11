package de

import (
	"fmt"
	"sync"
)

// DECounter tracks the number of used and pending Data Entries (DEs).
// It estimates the expected on-chain DE count, assuming all MsgSubmitDE
// transactions are eventually committed.
type DECounter struct {
	mu          sync.Mutex
	used        int64
	pending     int64
	blockHeight int64
}

// NewDECounter creates a new DECounter.
func NewDECounter() *DECounter {
	return &DECounter{
		mu:          sync.Mutex{},
		used:        0,
		pending:     0,
		blockHeight: 0,
	}
}

// AfterDEsCommitted updates internal counters after DEs are successfully committed.
func (dec *DECounter) AfterDEsCommitted(numDE int64) {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// pending cannot be negative as it represents the number of DEs
	// being created but not committed by the program
	dec.pending = max(0, dec.pending-numDE)

	// can be negative to represent that DEs on chain is more than expected
	dec.used -= numDE
}

// AfterDEsRejected updates internal counters after DEs are rejected.
func (dec *DECounter) AfterDEsRejected(numDE int64) {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// pending cannot be negative as it represents the number of DEs
	// being created but not committed by the program
	dec.pending = max(0, dec.pending-numDE)
}

// EvaluateDECreationFromUsage updates the internal counters based on the usage data.
// and return the number of DEs that needs to be created based on the given threshold.
func (dec *DECounter) EvaluateDECreationFromUsage(
	deUsed int64,
	threshold uint64,
	blockHeight int64,
) int64 {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// skip if the block height is less than the latest updated block height
	if dec.blockHeight >= blockHeight {
		return 0
	}

	dec.used += deUsed
	toBeCreated := dec.used - dec.pending
	if toBeCreated >= int64(threshold) {
		dec.pending += toBeCreated
		return toBeCreated
	}

	return 0
}

// AfterSyncWithChain syncs local state with actual on-chain DE count
// and updates internal counters.
func (dec *DECounter) AfterSyncWithChain(
	existing uint64,
	expectedDESize uint64,
	blockHeight int64,
) int64 {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	dec.blockHeight = blockHeight
	dec.used = int64(expectedDESize) - int64(existing)
	toBeCreated := max(0, dec.used-dec.pending)
	dec.pending += toBeCreated

	return toBeCreated
}

func (dec *DECounter) String() string {
	return fmt.Sprintf(
		"DECounter{used: %d, pending: %d, blockHeight: %d}",
		dec.used,
		dec.pending,
		dec.blockHeight,
	)
}
