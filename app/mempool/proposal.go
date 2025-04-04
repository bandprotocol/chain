package mempool

import (
	"fmt"
	"math"

	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxUint64 is the maximum value of a uint64.
const MaxUint64 = math.MaxUint64

// Proposal represents a block proposal under construction.
type Proposal struct {
	logger log.Logger

	// Txs is the list of transactions in the proposal.
	txs [][]byte
	// Cache helps quickly check for duplicates by tx hash.
	cache map[string]struct{}
	// maxBlockSpace is the maximum block space available for this proposal.
	maxBlockSpace BlockSpace
	// totalBlockSpace is the total block space used by the proposal.
	totalBlockSpace BlockSpace
}

// NewProposal returns a new empty proposal constrained by max block size and max gas limit.
func NewProposal(logger log.Logger, maxBlockSize int64, maxGasLimit uint64) Proposal {
	return Proposal{
		logger:          logger,
		txs:             make([][]byte, 0),
		cache:           make(map[string]struct{}),
		maxBlockSpace:   NewBlockSpace(maxBlockSize, maxGasLimit),
		totalBlockSpace: NewBlockSpace(0, 0),
	}
}

// Contains returns true if the proposal already has a transaction with the given txHash.
func (p *Proposal) Contains(txHash string) bool {
	_, ok := p.cache[txHash]
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
	p.cache[txInfo.Hash] = struct{}{}

	p.totalBlockSpace = currentBlockSpace

	return nil
}

// GetLimit returns the maximum block space available for a given ratio.
func (p *Proposal) GetLimit(ratio sdkmath.LegacyDec) BlockSpace {
	// In the case where the ratio is zero, we return the max tx bytes remaining.
	if ratio.IsZero() {
		return p.maxBlockSpace.Sub(p.totalBlockSpace)
	}

	return p.maxBlockSpace.MulDec(ratio)
}

// GetBlockLimits retrieves the maximum block size and gas limit from context.
func GetBlockLimits(ctx sdk.Context) (int64, uint64) {
	blockParams := ctx.ConsensusParams().Block

	var maxGasLimit uint64

	if blockParams.MaxGas == -1 {
		maxGasLimit = MaxUint64
	} else {
		maxGasLimit = uint64(blockParams.MaxGas)
	}

	maxBytesLimit := blockParams.MaxBytes
	if blockParams.MaxBytes == -1 {
		maxBytesLimit = comettypes.MaxBlockSizeBytes
	}

	return maxBytesLimit, maxGasLimit
}
