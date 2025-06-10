package de

import (
	"sync"
)

// DECounter tracks the number of used and pending DEs.
// It represents the expected on-chain DE count if all MsgSubmitDE
// transactions are successfully committed.
type DECounter struct {
	mu      sync.Mutex
	used    int64
	pending int64
}

// NewDECounter creates a new DECounter.
func NewDECounter() *DECounter {
	return &DECounter{
		mu:      sync.Mutex{},
		used:    0,
		pending: 0,
	}
}

// UpdateCommittedDEs updates the number of created DEs that being stored on chain.
// It decreases both the used and pending DE counts to reflect thatsome of the demand
// has been fulfilled.
func (dec *DECounter) UpdateCommittedDEs(numDE int64) {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// pending cannot be negative as it represents the number of DEs
	// being created but not committed by the program
	dec.pending = max(0, dec.pending-numDE)

	// can be negative to represent that DEs on chain is more than expected
	dec.used -= numDE
}

// UpdateRejectedDEs updates the number of rejected DEs. It decreases
// the number of pending DEs to reflect that the supply has been rejected
func (dec *DECounter) UpdateRejectedDEs(numDE int64) {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// pending cannot be negative as it represents the number of DEs
	// being created but not committed by the program
	dec.pending = max(0, dec.pending-numDE)
}

// CheckUsageAndAddPending checks if the sum of used and pending DEs is
// greater than the threshold and update the number of pending DEs if so.
// It returns the number of DEs that were added to the pending count.
func (dec *DECounter) CheckUsageAndAddPending(deUsed int64, threshold uint64) int64 {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	dec.used += deUsed
	toBeCreated := dec.used - dec.pending
	if toBeCreated >= int64(threshold) {
		dec.pending += toBeCreated
		return toBeCreated
	}

	return 0
}

// ComputeAndAddMissingDEs recalculates the number of used DEs and
// adds and return the missing DEs to the pending count if there is any.
func (dec *DECounter) ComputeAndAddMissingDEs(existing uint64, expectedDESize uint64) int64 {
	dec.mu.Lock()
	defer dec.mu.Unlock()

	// No need to update dec.used here — it will be updated later when a new signing is published
	// via CheckUsageAndAddPending. It's expected that dec.used will temporarily lag behind
	// the actual on-chain usage. the sum of dec.used and dec.pending will reflect
	// the expected DEs on chain in next CheckUsageAndAddPending.
	toBeCreated := max(0, int64(expectedDESize)-int64(existing)-dec.pending)
	dec.pending += toBeCreated

	return toBeCreated
}
