package band

import (
	"fmt"

	"github.com/skip-mev/block-sdk/v2/block/utils"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// ProposalHandler is a wrapper around the ABCI++ PrepareProposal and ProcessProposal
	// handlers for the BandMempool.
	ProposalHandler struct {
		logger                   log.Logger
		txDecoder                sdk.TxDecoder
		bandMempool              *BandMempool
		useCustomProcessProposal bool
	}
)

// NewDefaultProposalHandler returns a new ABCI++ proposal handler for the BandMempool.
// This proposal handler will not use custom process proposal logic.
func NewDefaultProposalHandler(
	logger log.Logger,
	txDecoder sdk.TxDecoder,
	bandMempool *BandMempool,
) *ProposalHandler {
	return &ProposalHandler{
		logger:                   logger,
		txDecoder:                txDecoder,
		bandMempool:              bandMempool,
		useCustomProcessProposal: false,
	}
}

// PrepareProposalHandler prepares the proposal by selecting transactions from the BandMempool.
func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
		if req.Height <= 1 {
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		// Recover from any panics during proposal preparation.
		defer func() {
			if rec := recover(); rec != nil {
				h.logger.Error("failed to prepare proposal", "err", err)
				resp = &abci.ResponsePrepareProposal{Txs: make([][]byte, 0)}
				err = fmt.Errorf("failed to prepare proposal: %v", rec)
			}
		}()

		h.logger.Info(
			"preparing proposal from BandMempool",
			"height", req.Height,
		)

		// Get the max gas limit and max block size for the proposal.
		_, maxGasLimit := GetBlockLimits(ctx)
		proposal := NewProposal(h.logger, req.MaxTxBytes, maxGasLimit)

		// Fill the proposal with transactions from the BandMempool.
		finalProposal, err := h.bandMempool.PrepareBandProposal(ctx, proposal)

		h.logger.Info(
			"prepared proposal",
			"num_txs", len(proposal.Txs),
			"total_tx_bytes", proposal.Info.BlockSize,
			"max_tx_bytes", proposal.Info.MaxBlockSize,
			"total_gas_limit", proposal.Info.GasLimit,
			"max_gas_limit", proposal.Info.MaxGasLimit,
			"height", req.Height,
		)

		return &abci.ResponsePrepareProposal{
			Txs: finalProposal.Txs,
		}, nil
	}
}

// ProcessProposalHandler processes the proposal by verifying all transactions in the proposal.
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	if !h.useCustomProcessProposal {
		return baseapp.NoOpProcessProposal()
	}

	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
		if req.Height <= 1 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		// Recover from any panics during proposal processing.
		defer func() {
			if rec := recover(); rec != nil {
				h.logger.Error("failed to process proposal", "recover_err", rec)
				resp = &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
				err = fmt.Errorf("failed to process proposal: %v", rec)
			}
		}()

		// Decode the transactions in the proposal.
		decodedTxs, err := utils.GetDecodedTxs(h.txDecoder, req.Txs)
		if err != nil {
			h.logger.Error("failed to decode txs", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// Verify each transaction in the proposal.
		for _, tx := range decodedTxs {
			// Perform custom verification logic here if needed.
			// For now, we assume all transactions are valid.
			h.logger.Info("verified transaction", "tx", tx)
		}

		h.logger.Info(
			"processed proposal",
			"num_txs", len(decodedTxs),
			"height", req.Height,
		)

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// ====================================
// Utils
// ====================================

const (
	// MaxUint64 is the maximum value of a uint64.
	MaxUint64 = 1<<64 - 1
)

// GetBlockLimits retrieves the maximum number of bytes and gas limit allowed in a block.
func GetBlockLimits(ctx sdk.Context) (int64, uint64) {
	blockParams := ctx.ConsensusParams().Block

	// If the max gas is set to 0, then the max gas limit for the block can be infinite.
	// Otherwise, we use the max gas limit casted as a uint64 which is how gas limits are
	// extracted from sdk.Tx's.
	var maxGasLimit uint64
	if maxGas := blockParams.MaxGas; maxGas > 0 {
		maxGasLimit = uint64(maxGas)
	} else {
		maxGasLimit = MaxUint64
	}

	return blockParams.MaxBytes, maxGasLimit
}

// ====================================
// Proposal
// ====================================

// ProposalInfo contains the metadata about a given proposal that was built by
// the block-sdk. This is used to verify and consolidate proposal data across
// the network.
type ProposalInfo struct {
	// TxsByLane contains information about how each partial proposal
	// was constructed by the block-sdk lanes.
	TxsByLane map[string]uint64 `protobuf:"bytes,1,rep,name=txs_by_lane,json=txsByLane,proto3"        json:"txs_by_lane,omitempty"    protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// MaxBlockSize corresponds to the upper bound on the size of the
	// block that was used to construct this block proposal.
	MaxBlockSize int64 `protobuf:"varint,2,opt,name=max_block_size,json=maxBlockSize,proto3" json:"max_block_size,omitempty"`
	// MaxGasLimit corresponds to the upper bound on the gas limit of the
	// block that was used to construct this block proposal.
	MaxGasLimit uint64 `protobuf:"varint,3,opt,name=max_gas_limit,json=maxGasLimit,proto3"   json:"max_gas_limit,omitempty"`
	// BlockSize corresponds to the size of this block proposal.
	BlockSize int64 `protobuf:"varint,4,opt,name=block_size,json=blockSize,proto3"        json:"block_size,omitempty"`
	// GasLimit corresponds to the gas limit of this block proposal.
	GasLimit uint64 `protobuf:"varint,5,opt,name=gas_limit,json=gasLimit,proto3"          json:"gas_limit,omitempty"`
}

type (
	// Proposal defines a block proposal type.
	Proposal struct {
		Logger log.Logger

		// Txs is the list of transactions in the proposal.
		Txs [][]byte
		// Cache is a cache of the selected transactions in the proposal.
		Cache map[string]struct{}
		// Info contains information about the state of the proposal.
		Info ProposalInfo

		currentSize int64
		currentGas  uint64
	}
)

// NewProposal returns a new empty proposal. Any transactions added to the proposal
// will be subject to the given max block size and max gas limit.
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

// Contains returns true if the proposal contains the given transaction.
func (p *Proposal) Contains(txHash string) bool {
	_, ok := p.Cache[txHash]
	return ok
}

// Add adds a transaction to the proposal.
func (p *Proposal) Add(txInfo TxWithInfo) error {
	fmt.Println("try add tx to proposal: ", txInfo.Hash)
	// if the transaction is already in the block proposal, return an error.
	if p.Contains(txInfo.Hash) {
		return fmt.Errorf("transaction already in proposal: %s", txInfo.Hash)
	}

	// if the transaction is too large, return an error.
	if p.currentSize+txInfo.Size > p.Info.MaxBlockSize {
		return fmt.Errorf(
			"transaction size exceeds max block size: %d > %d",
			p.currentSize+txInfo.Size,
			p.Info.MaxBlockSize,
		)
	}

	// if the transaction gas limit is too large, return an error.
	if p.currentGas+txInfo.GasLimit > p.Info.MaxGasLimit {
		return fmt.Errorf(
			"transaction gas limit exceeds max gas limit: %d > %d",
			p.currentGas+txInfo.GasLimit,
			p.Info.MaxGasLimit,
		)
	}

	// add the transaction to the proposal.
	p.Txs = append(p.Txs, txInfo.TxBytes)
	p.Cache[txInfo.Hash] = struct{}{}

	p.Info.BlockSize += txInfo.Size
	p.Info.GasLimit += txInfo.GasLimit

	p.currentSize += txInfo.Size
	p.currentGas += txInfo.GasLimit

	return nil
}
