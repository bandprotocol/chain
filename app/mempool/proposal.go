package mempool

import (
	"fmt"

	"cosmossdk.io/log"
)

// Proposal represents a block proposal under construction.
type Proposal struct {
	logger log.Logger

	// txs is the list of transactions in the proposal.
	txs [][]byte
	// seen helps quickly check for duplicates by tx hash.
	seen map[string]struct{}
	// maxBlockSpace is the maximum block space available for this proposal.
	maxBlockSpace BlockSpace
	// totalBlockSpaceUsed is the total block space used by the proposal.
	totalBlockSpaceUsed BlockSpace
}

// NewProposal returns a new empty proposal constrained by max block size and max gas limit.
func NewProposal(logger log.Logger, maxBlockSize uint64, maxGasLimit uint64) Proposal {
	return Proposal{
		logger:              logger,
		txs:                 make([][]byte, 0),
		seen:                make(map[string]struct{}),
		maxBlockSpace:       NewBlockSpace(maxBlockSize, maxGasLimit),
		totalBlockSpaceUsed: NewBlockSpace(0, 0),
	}
}

// Contains returns true if the proposal already has a transaction with the given txHash.
func (p *Proposal) Contains(txHash string) bool {
	_, ok := p.seen[txHash]
	return ok
}

// Add attempts to add a transaction to the proposal, respecting size/gas limits.
func (p *Proposal) Add(txInfo TxWithInfo) error {
	if p.Contains(txInfo.Hash) {
		return fmt.Errorf("transaction already in proposal: %s", txInfo.Hash)
	}

	currentBlockSpaceUsed := p.totalBlockSpaceUsed.Add(txInfo.BlockSpace)

	// Check block size limit
	if p.maxBlockSpace.IsExceededBy(currentBlockSpaceUsed) {
		return fmt.Errorf(
			"transaction space exceeds max block space: %s > %s",
			currentBlockSpaceUsed.String(), p.maxBlockSpace.String(),
		)
	}

	// Add transaction
	p.txs = append(p.txs, txInfo.TxBytes)
	p.seen[txInfo.Hash] = struct{}{}

	p.totalBlockSpaceUsed = currentBlockSpaceUsed

	return nil
}

// GetRemainingBlockSpace returns the remaining block space available for the proposal.
func (p *Proposal) GetRemainingBlockSpace() BlockSpace {
	return p.maxBlockSpace.Sub(p.totalBlockSpaceUsed)
}
