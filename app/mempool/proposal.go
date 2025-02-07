package mempool

import (
	"fmt"

	"cosmossdk.io/log"

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

// GetBlockLimits retrieves the maximum block size and gas limit from context.
func GetBlockLimits(ctx sdk.Context) (int64, uint64) {
	blockParams := ctx.ConsensusParams().Block

	var maxGasLimit uint64
	if maxGas := blockParams.MaxGas; maxGas > 0 {
		maxGasLimit = uint64(maxGas)
	} else {
		maxGasLimit = MaxUint64
	}

	return blockParams.MaxBytes, maxGasLimit
}
