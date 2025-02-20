package mempool

import (
	"fmt"

	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxUint64 is the maximum value of a uint64.
const MaxUint64 = 1<<64 - 1

// Proposal represents a block proposal under construction.
type Proposal struct {
	Logger log.Logger

	// Txs is the list of transactions in the proposal.
	Txs [][]byte
	// Cache helps quickly check for duplicates by tx hash.
	Cache map[string]struct{}
	// MaxBlockSpace is the maximum block space available for this proposal.
	MaxBlockSpace BlockSpace
	// TotalBlockSpace is the total block space used by the proposal.
	TotalBlockSpace BlockSpace
}

// NewProposal returns a new empty proposal constrained by max block size and max gas limit.
func NewProposal(logger log.Logger, maxBlockSize int64, maxGasLimit uint64) Proposal {
	return Proposal{
		Logger:          logger,
		Txs:             make([][]byte, 0),
		Cache:           make(map[string]struct{}),
		MaxBlockSpace:   NewBlockSpace(maxBlockSize, maxGasLimit),
		TotalBlockSpace: NewBlockSpace(0, 0),
	}
}

// Contains returns true if the proposal already has a transaction with the given txHash.
func (p *Proposal) Contains(txHash string) bool {
	_, ok := p.Cache[txHash]
	return ok
}

// Add attempts to add a transaction to the proposal, respecting size/gas limits.
func (p *Proposal) Add(txInfo TxWithInfo) error {
	fmt.Println("try add tx to proposal", txInfo.Hash)
	if p.Contains(txInfo.Hash) {
		return fmt.Errorf("transaction already in proposal: %s", txInfo.Hash)
	}

	currentBlockSpace := p.TotalBlockSpace.Add(txInfo.BlockSpace)

	// Check block size limit
	if p.MaxBlockSpace.IsExceededBy(currentBlockSpace) {
		return fmt.Errorf(
			"transaction space exceeds max block space: %s > %s",
			currentBlockSpace.String(), p.MaxBlockSpace.String(),
		)
	}

	// Add transaction
	p.Txs = append(p.Txs, txInfo.TxBytes)
	p.Cache[txInfo.Hash] = struct{}{}

	p.TotalBlockSpace = currentBlockSpace

	return nil
}

// GetLimit returns the maximum block space available for a given ratio.
func (p *Proposal) GetLimit(ratio math.LegacyDec) BlockSpace {
	// In the case where the ratio is zero, we return the max tx bytes remaining.
	if ratio.IsZero() {
		return p.MaxBlockSpace.Sub(p.TotalBlockSpace)
	}

	return p.MaxBlockSpace.MulDec(ratio)
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
