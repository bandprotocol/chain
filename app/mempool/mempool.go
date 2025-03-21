package mempool

import (
	"context"
	"fmt"

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

	cacheSdkCtx, _ := sdk.UnwrapSDKContext(ctx).CacheContext()
	for _, lane := range m.lanes {
		if lane.Match(cacheSdkCtx, tx) {
			return lane.Insert(ctx, tx)
		}
	}

	return nil
}

// Select returns a Mempool iterator (currently nil).
func (m *Mempool) Select(ctx context.Context, txs [][]byte) sdkmempool.Iterator {
	return nil
}

// CountTx returns the total number of transactions across all lanes.
func (m *Mempool) CountTx() int {
	count := 0
	for _, lane := range m.lanes {
		count += lane.CountTx()
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
// then calls each lane’s FillProposal method. If leftover gas is important to you,
// you can implement a second pass or distribute leftover to subsequent lanes, etc.
func (m *Mempool) PrepareProposal(ctx sdk.Context, proposal Proposal) (Proposal, error) {
	cacheCtx, _ := ctx.CacheContext()

	// 1) Perform the initial fill of proposals
	laneIterators, txsToRemove, blockUsed := m.fillInitialProposals(cacheCtx, &proposal)

	// 2) Calculate the remaining block space
	remainderLimit := proposal.MaxBlockSpace.Sub(blockUsed)

	// 3) Fill proposals with leftover space
	m.fillRemainderProposals(&proposal, laneIterators, txsToRemove, remainderLimit)

	// 4) Remove the transactions that were invalidated from each lane
	m.removeTxsFromLanes(txsToRemove)

	return proposal, nil
}

// fillInitialProposals iterates over lanes, calling FillProposal. It returns:
//   - laneIterators:  the Iterator for each lane
//   - txsToRemove:    slice-of-slice of transactions to be removed later
//   - totalSize:      total block size used
//   - totalGas:       total gas used
func (m *Mempool) fillInitialProposals(
	ctx sdk.Context,
	proposal *Proposal,
) (
	[]sdkmempool.Iterator,
	[][]sdk.Tx,
	BlockSpace,
) {
	totalBlockUsed := NewBlockSpace(0, 0)

	laneIterators := make([]sdkmempool.Iterator, len(m.lanes))
	txsToRemove := make([][]sdk.Tx, len(m.lanes))

	for i, lane := range m.lanes {
		blockUsed, iterator, txs := lane.FillProposal(ctx, proposal)
		totalBlockUsed.IncreaseBy(blockUsed)

		laneIterators[i] = iterator
		txsToRemove[i] = txs
	}

	return laneIterators, txsToRemove, totalBlockUsed
}

// fillRemainderProposals performs an additional fill on each lane using the leftover
// BlockSpace. It updates txsToRemove to include any newly removed transactions.
func (m *Mempool) fillRemainderProposals(
	proposal *Proposal,
	laneIterators []sdkmempool.Iterator,
	txsToRemove [][]sdk.Tx,
	remainderLimit BlockSpace,
) {
	for i, lane := range m.lanes {
		blockUsed, removedTxs := lane.FillProposalBy(
			proposal,
			laneIterators[i],
			remainderLimit,
		)

		// Decrement the remainder for subsequent lanes
		remainderLimit.DecreaseBy(blockUsed)

		// Append any newly removed transactions to be removed
		txsToRemove[i] = append(txsToRemove[i], removedTxs...)
	}
}

// removeTxsFromLanes loops through each lane and removes all transactions
// accumulated in txsToRemove.
func (m *Mempool) removeTxsFromLanes(txsToRemove [][]sdk.Tx) {
	for i, lane := range m.lanes {
		for _, tx := range txsToRemove[i] {
			if err := lane.Remove(tx); err != nil {
				m.logger.Error(
					"failed to remove transactions from lane",
					"lane", lane.Name,
					"err", err,
				)
			}
		}
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
