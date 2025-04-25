package mempool

import (
	"fmt"
	"math"

	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// totalBlockSpace is the total block space used by the proposal.
	totalBlockSpace BlockSpace
}

// NewProposal returns a new empty proposal constrained by max block size and max gas limit.
func NewProposal(logger log.Logger, maxBlockSize uint64, maxGasLimit uint64) Proposal {
	return Proposal{
		logger:          logger,
		txs:             make([][]byte, 0),
		seen:            make(map[string]struct{}),
		maxBlockSpace:   NewBlockSpace(maxBlockSize, maxGasLimit),
		totalBlockSpace: NewBlockSpace(0, 0),
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

	currentBlockSpace := p.totalBlockSpace.Add(txInfo.BlockSpace)

	// Check block size limit
	if p.maxBlockSpace.IsExceededBy(currentBlockSpace) {
		return fmt.Errorf(
			"transaction space exceeds max block space: %s > %s",
			currentBlockSpace.String(), p.maxBlockSpace.String(),
		)
	}

	// Add transaction
	p.txs = append(p.txs, txInfo.TxBytes)
	p.seen[txInfo.Hash] = struct{}{}

	p.totalBlockSpace = currentBlockSpace

	return nil
}

// GetBlockLimits retrieves the maximum block size and gas limit from context.
func GetBlockLimits(ctx sdk.Context) (uint64, uint64) {
	blockParams := ctx.ConsensusParams().Block

	var maxBytesLimit uint64
	if blockParams.MaxBytes == -1 {
		maxBytesLimit = uint64(comettypes.MaxBlockSizeBytes)
	} else {
		maxBytesLimit = uint64(blockParams.MaxBytes)
	}

	var maxGasLimit uint64
	if blockParams.MaxGas == -1 {
		maxGasLimit = math.MaxUint64
	} else {
		maxGasLimit = uint64(blockParams.MaxGas)
	}

	return maxBytesLimit, maxGasLimit
}
