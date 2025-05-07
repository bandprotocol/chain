package mempool

import (
	"context"
	"fmt"
	"math"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

var _ sdkmempool.Mempool = (*Mempool)(nil)

// Mempool implements the sdkmempool.Mempool interface and uses Lanes internally.
type Mempool struct {
	logger log.Logger

	lanes []*Lane
}

// NewMempool returns a new mempool with the given lanes.
func NewMempool(
	logger log.Logger,
	lanes []*Lane,
) *Mempool {
	return &Mempool{
		logger: logger,
		lanes:  lanes,
	}
}

// Insert will insert a transaction into the mempool. It inserts the transaction
// into the first lane that it matches.
func (m *Mempool) Insert(ctx context.Context, tx sdk.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("panic in Insert", "err", r)
			err = fmt.Errorf("panic in Insert: %v", r)
		}
	}()

	cacheSDKCtx, _ := sdk.UnwrapSDKContext(ctx).CacheContext()
	for _, lane := range m.lanes {
		if lane.Match(cacheSDKCtx, tx) {
			return lane.Insert(ctx, tx)
		}
	}

	return
}

// Select returns a Mempool iterator (currently nil).
func (m *Mempool) Select(ctx context.Context, txs [][]byte) sdkmempool.Iterator {
	return nil
}

// CountTx returns the total number of transactions across all lanes.
// Returns math.MaxInt if the total count would overflow.
func (m *Mempool) CountTx() int {
	count := 0
	for _, lane := range m.lanes {
		laneCount := lane.CountTx()
		if laneCount > 0 && count > math.MaxInt-laneCount {
			// If adding laneCount would cause overflow, return MaxInt
			return math.MaxInt
		}
		count += laneCount
	}
	return count
}

// Remove removes a transaction from the mempool. This assumes that the transaction
// is contained in only one of the lanes.
func (m *Mempool) Remove(tx sdk.Tx) (err error) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("panic in Remove", "err", r)
			err = fmt.Errorf("panic in Remove: %v", r)
		}
	}()

	for _, lane := range m.lanes {
		if lane.Contains(tx) {
			return lane.Remove(tx)
		}
	}

	return nil
}

// PrepareProposal divides the block gas limit among lanes (based on lane percentage),
// then calls each lane's FillProposal method. If leftover gas is important to you,
// you can implement a second pass or distribute leftover to subsequent lanes, etc.
func (m *Mempool) PrepareProposal(ctx sdk.Context, proposal Proposal) (Proposal, error) {
	cacheCtx, _ := ctx.CacheContext()

	// 1) Perform the initial fill of proposals
	laneIterators, blockUsed := m.fillInitialProposals(cacheCtx, &proposal)

	// 2) Calculate the remaining block space
	remainderLimit := proposal.maxBlockSpace.Sub(blockUsed)

	// 3) Fill proposals with leftover space
	m.fillRemainderProposals(&proposal, laneIterators, remainderLimit)

	return proposal, nil
}

// fillInitialProposals iterates over lanes, calling FillProposal. It returns:
//   - laneIterators:  the Iterator for each lane
//   - totalSize:      total block size used
//   - totalGas:       total gas used
func (m *Mempool) fillInitialProposals(
	ctx sdk.Context,
	proposal *Proposal,
) (
	[]sdkmempool.Iterator,
	BlockSpace,
) {
	totalBlockUsed := NewBlockSpace(0, 0)

	laneIterators := make([]sdkmempool.Iterator, len(m.lanes))

	for i, lane := range m.lanes {
		blockUsed, iterator := lane.FillProposal(ctx, proposal)
		totalBlockUsed = totalBlockUsed.Add(blockUsed)

		laneIterators[i] = iterator
	}

	return laneIterators, totalBlockUsed
}

// fillRemainderProposals performs an additional fill on each lane using the leftover
// BlockSpace. It updates txsToRemove to include any newly removed transactions.
func (m *Mempool) fillRemainderProposals(
	proposal *Proposal,
	laneIterators []sdkmempool.Iterator,
	remainderLimit BlockSpace,
) {
	for i, lane := range m.lanes {
		blockUsed := lane.FillProposalByIterator(
			proposal,
			laneIterators[i],
			remainderLimit,
		)

		// Decrement the remainder for subsequent lanes
		remainderLimit = remainderLimit.Sub(blockUsed)
	}
}

// Contains returns true if the transaction is contained in any of the lanes.
func (m *Mempool) Contains(tx sdk.Tx) (contains bool) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("panic in Contains", "err", r)
			contains = false
		}
	}()

	for _, lane := range m.lanes {
		if lane.Contains(tx) {
			return true
		}
	}

	return false
}

// GetLane returns the lane with the given name.
func (m *Mempool) GetLane(name string) *Lane {
	for _, lane := range m.lanes {
		if lane.name == name {
			return lane
		}
	}

	return nil
}
