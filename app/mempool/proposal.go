package mempool

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	comettypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxUint64 is the maximum value of a uint64.
const MaxUint64 = 1<<64 - 1

// ProposalInfo contains metadata about how this proposal was constructed.
type ProposalInfo struct {
	// TxsByLane is a map counting how many transactions came from each "lane" (optional usage).
	TxsByLane map[string]uint64 `json:"txs_by_lane,omitempty"`

	// MaxBlockSize is the upper bound on the block size used to construct this proposal.
	MaxBlockSize int64 `json:"max_block_size,omitempty"`
	// MaxGasLimit is the upper bound on the total gas used to construct this proposal.
	MaxGasLimit uint64 `json:"max_gas_limit,omitempty"`

	// BlockSize is the current total block size of this proposal.
	BlockSize int64 `json:"block_size,omitempty"`
	// GasLimit is the current total gas of this proposal.
	GasLimit uint64 `json:"gas_limit,omitempty"`
}

// Proposal represents a block proposal under construction.
type Proposal struct {
	Logger log.Logger

	// Txs is the list of transactions in the proposal.
	Txs [][]byte
	// Cache helps quickly check for duplicates by tx hash.
	Cache map[string]struct{}
	// Info contains metadata about the proposal's block usage.
	Info ProposalInfo

	currentSize int64
	currentGas  uint64
}

// NewProposal returns a new empty proposal constrained by max block size and max gas limit.
func NewProposal(logger log.Logger, maxBlockSize int64, maxGasLimit uint64) Proposal {
	return Proposal{
		Logger: logger,
		Txs:    make([][]byte, 0),
		Cache:  make(map[string]struct{}),
		Info: ProposalInfo{
			TxsByLane:    make(map[string]uint64),
			MaxBlockSize: maxBlockSize,
			MaxGasLimit:  maxGasLimit,
		},
	}
}

// Contains returns true if the proposal already has a transaction with the given txHash.
func (p *Proposal) Contains(txHash string) bool {
	_, ok := p.Cache[txHash]
	return ok
}

// Add attempts to add a transaction to the proposal, respecting size/gas limits.
func (p *Proposal) Add(txInfo TxWithInfo) error {
	if p.Contains(txInfo.Hash) {
		return fmt.Errorf("transaction already in proposal: %s", txInfo.Hash)
	}

	// Check block size limit
	if p.currentSize+txInfo.Size > p.Info.MaxBlockSize {
		return fmt.Errorf(
			"transaction size exceeds max block size: %d > %d",
			p.currentSize+txInfo.Size,
			p.Info.MaxBlockSize,
		)
	}

	// Check block gas limit
	if p.currentGas+txInfo.GasLimit > p.Info.MaxGasLimit {
		return fmt.Errorf(
			"transaction gas limit exceeds max gas limit: %d > %d",
			p.currentGas+txInfo.GasLimit,
			p.Info.MaxGasLimit,
		)
	}

	// Add transaction
	p.Txs = append(p.Txs, txInfo.TxBytes)
	p.Cache[txInfo.Hash] = struct{}{}

	p.Info.BlockSize += txInfo.Size
	p.Info.GasLimit += txInfo.GasLimit

	p.currentSize += txInfo.Size
	p.currentGas += txInfo.GasLimit

	return nil
}

// GetLimit returns the maximum block space available for a given ratio.
func (p *Proposal) GetLimit(ratio math.LegacyDec) BlockSpace {
	var (
		txBytesLimit int64
		gasLimit     uint64
	)

	// In the case where the ratio is zero, we return the max tx bytes remaining.
	if ratio.IsZero() {
		txBytesLimit = p.Info.MaxBlockSize - p.Info.BlockSize
		if txBytesLimit < 0 {
			txBytesLimit = 0
		}

		// Unsigned subtraction needs an additional check
		if p.Info.GasLimit >= p.Info.MaxGasLimit {
			gasLimit = 0
		} else {
			gasLimit = p.Info.MaxGasLimit - p.Info.GasLimit
		}
	} else {
		// Otherwise, we calculate the max tx bytes / gas limit for the lane based on the ratio.
		txBytesLimit = ratio.MulInt64(p.Info.MaxBlockSize).TruncateInt().Int64()
		gasLimit = ratio.MulInt(math.NewIntFromUint64(p.Info.MaxGasLimit)).TruncateInt().Uint64()
	}

	return NewBlockSpace(txBytesLimit, gasLimit)
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
