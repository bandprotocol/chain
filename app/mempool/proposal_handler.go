package mempool

import (
	"fmt"
	"math"

	abci "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProposalHandler wraps ABCI++ PrepareProposal/ProcessProposal for the Mempool.
type ProposalHandler struct {
	logger    log.Logger
	txDecoder sdk.TxDecoder
	mempool   *Mempool
}

// NewProposalHandler returns a new ABCI++ proposal handler for the Mempool.
func NewProposalHandler(
	logger log.Logger,
	txDecoder sdk.TxDecoder,
	mempool *Mempool,
) *ProposalHandler {
	return &ProposalHandler{
		logger:    logger,
		txDecoder: txDecoder,
		mempool:   mempool,
	}
}

// PrepareProposalHandler builds the next block proposal from the Mempool.
func (h *ProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
		// For height <= 1, just return the default TXs (e.g., chain start).
		if req.Height <= 1 {
			return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
		}

		defer func() {
			if rec := recover(); rec != nil {
				h.logger.Error("failed to prepare proposal", "err", err)
				resp = &abci.ResponsePrepareProposal{Txs: make([][]byte, 0)}
				err = fmt.Errorf("failed to prepare proposal: %v", rec)
			}
		}()

		h.logger.Info("preparing proposal from Mempool", "height", req.Height)

		// Gather block limits
		maxBytesLimit, maxGasLimit := getBlockLimits(ctx)
		var maxTxBytes uint64
		if req.MaxTxBytes < 0 {
			maxTxBytes = maxBytesLimit
		} else {
			maxTxBytes = uint64(req.MaxTxBytes)
		}

		proposal := NewProposal(
			h.logger,
			min(maxTxBytes, maxBytesLimit),
			maxGasLimit,
		)

		// Populate proposal from Mempool
		finalProposal, err := h.mempool.PrepareProposal(ctx, proposal)
		if err != nil {
			// If an error occurs, we can still return what we have or choose to return nothing
			h.logger.Error("failed to prepare proposal", "err", err)
			return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, err
		}

		h.logger.Info(
			"prepared proposal",
			"num_txs", len(finalProposal.txs),
			"total_block_space", finalProposal.totalBlockSpaceUsed.String(),
			"max_block_space", finalProposal.maxBlockSpace.String(),
			"height", req.Height,
		)

		return &abci.ResponsePrepareProposal{
			Txs: finalProposal.txs,
		}, nil
	}
}

// ProcessProposalHandler returns a no-op process proposal handler.
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return baseapp.NoOpProcessProposal()
}

// getBlockLimits retrieves the maximum block size and gas limit from context.
func getBlockLimits(ctx sdk.Context) (uint64, uint64) {
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
