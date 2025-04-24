package mempool

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

// Lane defines a logical grouping of transactions within the mempool.
type Lane struct {
	logger          log.Logger
	txEncoder       sdk.TxEncoder
	signerExtractor sdkmempool.SignerExtractionAdapter
	name            string
	matchFn         func(ctx sdk.Context, tx sdk.Tx) bool

	maxTransactionBlockRatio math.LegacyDec
	maxLaneBlockRatio        math.LegacyDec

	laneMempool sdkmempool.Mempool

	// txIndex holds the uppercase hex-encoded hash for every transaction
	// currently in this lane's mempool.
	txIndex map[string]struct{}

	// callbackAfterFillProposal is a callback function that is called after
	// filling the proposal with transactions from the lane.
	callbackAfterFillProposal func(isLaneLimitExceeded bool)

	// blocked indicates whether the transactions in this lane should be
	// excluded from the proposal for the current block.
	blocked bool

	// Add mutex for thread safety.
	mu sync.RWMutex
}

// NewLane is a constructor for a lane.
func NewLane(
	logger log.Logger,
	txEncoder sdk.TxEncoder,
	signerExtractor sdkmempool.SignerExtractionAdapter,
	name string,
	matchFn func(sdk.Context, sdk.Tx) bool,
	maxTransactionBlockRatio math.LegacyDec,
	maxLaneBlockRatio math.LegacyDec,
	laneMempool sdkmempool.Mempool,
	callbackAfterFillProposal func(isLaneLimitExceeded bool),
) *Lane {
	return &Lane{
		logger:                    logger,
		txEncoder:                 txEncoder,
		signerExtractor:           signerExtractor,
		name:                      name,
		matchFn:                   matchFn,
		maxTransactionBlockRatio:  maxTransactionBlockRatio,
		maxLaneBlockRatio:         maxLaneBlockRatio,
		laneMempool:               laneMempool,
		callbackAfterFillProposal: callbackAfterFillProposal,

		// Initialize the txIndex.
		txIndex: make(map[string]struct{}),

		blocked: false,
	}
}

// Insert inserts a transaction into the lane's mempool.
func (l *Lane) Insert(ctx context.Context, tx sdk.Tx) error {
	txInfo, err := l.getTxInfo(tx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	consensusParams := sdkCtx.ConsensusParams()
	transactionLimit := NewBlockSpace(
		uint64(consensusParams.Block.MaxBytes),
		uint64(consensusParams.Block.MaxGas),
	).Scale(l.maxTransactionBlockRatio)

	if transactionLimit.IsExceededBy(txInfo.BlockSpace) {
		return fmt.Errorf(
			"transaction exceeds limit: tx_hash %s, lane %s, limit %s, tx_size %s",
			txInfo.Hash,
			l.name,
			transactionLimit,
			txInfo.BlockSpace,
		)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err = l.laneMempool.Insert(ctx, tx); err != nil {
		return err
	}

	l.txIndex[txInfo.Hash] = struct{}{}
	return nil
}

// CountTx returns the total number of transactions in the lane's mempool.
func (l *Lane) CountTx() int {
	return l.laneMempool.CountTx()
}

// Remove removes a transaction from the lane's mempool.
func (l *Lane) Remove(tx sdk.Tx) error {
	txInfo, err := l.getTxInfo(tx)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err = l.laneMempool.Remove(tx); err != nil {
		return err
	}

	delete(l.txIndex, txInfo.Hash)
	return nil
}

// Contains returns true if the lane's mempool contains the transaction.
func (l *Lane) Contains(tx sdk.Tx) bool {
	txInfo, err := l.getTxInfo(tx)
	if err != nil {
		return false
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	_, exists := l.txIndex[txInfo.Hash]
	return exists
}

// Match returns true if the transaction belongs to the lane.
func (l *Lane) Match(ctx sdk.Context, tx sdk.Tx) bool {
	return l.matchFn(ctx, tx)
}

// FillProposal fills the proposal with transactions from the lane mempool with its own limit.
// It returns the total size and gas of the transactions added to the proposal.
// It also returns an iterator to the next transaction in the mempool.
func (l *Lane) FillProposal(
	ctx sdk.Context,
	proposal *Proposal,
) (blockUsed BlockSpace, iterator sdkmempool.Iterator) {
	// if the lane is blocked, we do not add any transactions to the proposal.
	if l.blocked {
		l.logger.Info("lane %s is blocked, skipping proposal filling", l.name)
		return
	}

	// Get the lane limit for the lane.
	laneLimit := proposal.maxBlockSpace.Scale(l.maxLaneBlockRatio)

	// Select all transactions in the mempool that are valid and not already in the
	// partial proposal.
	for iterator = l.laneMempool.Select(ctx, nil); iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReachedBy(blockUsed) {
			break
		}

		tx := iterator.Tx()
		txInfo, err := l.getTxInfo(tx)
		if err != nil {
			// If the transaction is not valid, we skip it.
			// This should never happen, but we log it for debugging purposes.
			l.logger.Error("failed to get tx info", "err", err)
			continue
		}

		// Add the transaction to the proposal.
		if err := proposal.Add(txInfo); err != nil {
			l.logger.Info(
				"failed to add tx to proposal",
				"lane", l.name,
				"tx_hash", txInfo.Hash,
				"err", err,
			)

			break
		}

		blockUsed = blockUsed.Add(txInfo.BlockSpace)
	}

	// call the callback function of the lane after fill proposal.
	if l.callbackAfterFillProposal != nil {
		l.callbackAfterFillProposal(laneLimit.IsReachedBy(blockUsed))
	}

	return
}

// FillProposalByIterator fills the proposal with transactions from the lane mempool with the given iterator and limit.
// It returns the total size and gas of the transactions added to the proposal.
func (l *Lane) FillProposalByIterator(
	proposal *Proposal,
	iterator sdkmempool.Iterator,
	laneLimit BlockSpace,
) (blockUsed BlockSpace) {
	// if the lane is blocked, we do not add any transactions to the proposal.
	if l.blocked {
		return
	}

	// Select all transactions in the mempool that are valid and not already in the partial proposal.
	for ; iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReachedBy(blockUsed) {
			break
		}

		tx := iterator.Tx()
		txInfo, err := l.getTxInfo(tx)
		if err != nil {
			// If the transaction is not valid, we skip it.
			// This should never happen, but we log it for debugging purposes.
			l.logger.Error("failed to get tx info", "err", err)
			continue
		}

		// Add the transaction to the proposal.
		if err := proposal.Add(txInfo); err != nil {
			l.logger.Info(
				"failed to add tx to proposal",
				"lane", l.name,
				"tx_hash", txInfo.Hash,
				"err", err,
			)

			break
		}

		// Update the total size and gas.
		blockUsed = blockUsed.Add(txInfo.BlockSpace)
	}

	return
}

// getTxInfo returns various information about the transaction that
// belongs to the lane including its priority, signer's, sequence number,
// size and more.
func (l *Lane) getTxInfo(tx sdk.Tx) (TxWithInfo, error) {
	txBytes, err := l.txEncoder(tx)
	if err != nil {
		return TxWithInfo{}, fmt.Errorf("failed to encode transaction: %w", err)
	}

	gasTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return TxWithInfo{}, fmt.Errorf("failed to cast transaction to gas tx")
	}

	signers, err := l.signerExtractor.GetSigners(tx)
	if err != nil {
		return TxWithInfo{}, err
	}

	blockSpace := NewBlockSpace(uint64(len(txBytes)), gasTx.GetGas())

	return TxWithInfo{
		Hash:       strings.ToUpper(hex.EncodeToString(comettypes.Tx(txBytes).Hash())),
		BlockSpace: blockSpace,
		TxBytes:    txBytes,
		Signers:    signers,
	}, nil
}

// SetBlocked sets the blocked flag to the given value.
func (l *Lane) SetBlocked(blocked bool) {
	l.blocked = blocked
}
