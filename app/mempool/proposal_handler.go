package mempool

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// ProposalHandler wraps ABCI++ PrepareProposal/ProcessProposal for the Mempool.
	ProposalHandler struct {
		logger                   log.Logger
		txDecoder                sdk.TxDecoder
		Mempool                  *Mempool
		useCustomProcessProposal bool
	}
)

// NewProposalHandler returns a new ABCI++ proposal handler for the Mempool.
func NewProposalHandler(
	logger log.Logger,
	txDecoder sdk.TxDecoder,
	mempool *Mempool,
) *ProposalHandler {
	return &ProposalHandler{
		logger:                   logger,
		txDecoder:                txDecoder,
		Mempool:                  mempool,
		useCustomProcessProposal: false, // set to true if you want custom logic
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
		_, maxGasLimit := GetBlockLimits(ctx)
		proposal := NewProposal(h.logger, req.MaxTxBytes, maxGasLimit)

		// Populate proposal from Mempool
		finalProposal, err := h.Mempool.PrepareProposal(ctx, proposal)
		if err != nil {
			// If an error occurs, we can still return what we have or choose to return nothing
			h.logger.Error("failed to prepare  proposal", "err", err)
			return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, err
		}

		h.logger.Info(
			"prepared proposal",
			"num_txs", len(finalProposal.Txs),
			"total_block_space", finalProposal.TotalBlockSpace.String(),
			"max_block_space", finalProposal.MaxBlockSpace.String(),
			"height", req.Height,
		)

		return &abci.ResponsePrepareProposal{
			Txs: finalProposal.Txs,
		}, nil
	}
}

// ProcessProposalHandler returns a no-op process proposal handler.
func (h *ProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return baseapp.NoOpProcessProposal()
}
