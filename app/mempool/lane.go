package mempool

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	comettypes "github.com/cometbft/cometbft/types"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

// Lane defines a logical grouping of transactions within the mempool.
type Lane struct {
	Logger          log.Logger
	TxEncoder       sdk.TxEncoder
	SignerExtractor sdkmempool.SignerExtractionAdapter
	Name            string
	Match           func(ctx sdk.Context, tx sdk.Tx) bool

	MaxTransactionSpace math.LegacyDec
	MaxLaneSpace        math.LegacyDec

	laneMempool sdkmempool.Mempool

	// txIndex holds the uppercase hex-encoded hash for every transaction
	// currently in this lane's mempool.
	txIndex map[string]struct{}

	// OnFilled is a callback function that is called when the lane exceeds its limit.
	OnFilled func(isFilled bool)

	isBlocked bool
}

// NewLane is a constructor for a lane.
func NewLane(
	logger log.Logger,
	txEncoder sdk.TxEncoder,
	signerExtractor sdkmempool.SignerExtractionAdapter,
	name string,
	matchFn func(sdk.Context, sdk.Tx) bool,
	maxTransactionSpace math.LegacyDec,
	maxLaneSpace math.LegacyDec,
	laneMempool sdkmempool.Mempool,
	onFilled func(isFilled bool),
) *Lane {
	return &Lane{
		Logger:              logger,
		TxEncoder:           txEncoder,
		SignerExtractor:     signerExtractor,
		Name:                name,
		Match:               matchFn,
		MaxTransactionSpace: maxTransactionSpace,
		MaxLaneSpace:        maxLaneSpace,
		laneMempool:         laneMempool,
		OnFilled:            onFilled,

		// Initialize the txIndex.
		txIndex: make(map[string]struct{}),

		isBlocked: false,
	}
}

// Insert inserts a transaction into the lane's mempool.
func (l *Lane) Insert(ctx context.Context, tx sdk.Tx) error {
	txInfo, err := l.GetTxInfo(tx)
	if err != nil {
		return err
	}

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
	txInfo, err := l.GetTxInfo(tx)
	if err != nil {
		return err
	}

	if err = l.laneMempool.Remove(tx); err != nil {
		return err
	}

	// Remove it from the local index
	delete(l.txIndex, txInfo.Hash)
	return nil
}

// Contains returns true if the lane's mempool contains the transaction.
func (l *Lane) Contains(tx sdk.Tx) bool {
	txInfo, err := l.GetTxInfo(tx)
	if err != nil {
		return false
	}

	_, exists := l.txIndex[txInfo.Hash]
	return exists
}

// FillProposal fills the proposal with transactions from the lane mempool with the its own limit.
// It returns the total size and gas of the transactions added to the proposal.
// It also returns an iterator to the next transaction in the mempool and a list
// of transactions that were removed from the lane mempool.
func (l *Lane) FillProposal(
	ctx sdk.Context,
	proposal *Proposal,
) (blockUsed BlockSpace, iterator sdkmempool.Iterator, txsToRemove []sdk.Tx) {
	var (
		transactionLimit BlockSpace
		laneLimit        BlockSpace
	)
	// Get the transaction and lane limit for the lane.
	transactionLimit = proposal.GetLimit(l.MaxTransactionSpace)
	laneLimit = proposal.GetLimit(l.MaxLaneSpace)

	isFilled := false

	if l.isBlocked {
		return
	}

	// Select all transactions in the mempool that are valid and not already in the
	// partial proposal.
	for iterator = l.laneMempool.Select(ctx, nil); iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReachedBy(blockUsed) {
			break
		}

		tx := iterator.Tx()
		txInfo, err := l.GetTxInfo(tx)
		if err != nil {
			l.Logger.Info("failed to get hash of tx", "err", err)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// if the transaction is exceed the limit, we remove it from the lane mempool.
		if transactionLimit.IsExceededBy(txInfo.BlockSpace) {
			l.Logger.Info(
				"failed to select tx for lane; tx exceeds limit",
				"tx_hash", txInfo.Hash,
				"lane", l.Name,
			)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// Add the transaction to the proposal.
		// TODO: check if the transaction cannot be added here, it should also cannot be added afterward.
		if err := proposal.Add(txInfo); err != nil {
			l.Logger.Info(
				"failed to add tx to proposal",
				"lane", l.Name,
				"tx_hash", txInfo.Hash,
				"err", err,
			)

			break
		}

		blockUsed.IncreaseBy(txInfo.BlockSpace)
	}

	if laneLimit.IsReachedBy(blockUsed) {
		isFilled = true
	}

	if l.OnFilled != nil {
		l.OnFilled(isFilled)
	}

	return
}

// FillProposalBy fills the proposal with transactions from the lane mempool with the given iterator and limit.
// It returns the total size and gas of the transactions added to the proposal.
// It also returns a list of transactions that were removed from the lane mempool.
func (l *Lane) FillProposalBy(
	proposal *Proposal,
	iterator sdkmempool.Iterator,
	laneLimit BlockSpace,
) (blockUsed BlockSpace, txsToRemove []sdk.Tx) {
	// get the transaction limit for the lane.
	transactionLimit := proposal.GetLimit(l.MaxTransactionSpace)

	if l.isBlocked {
		return
	}

	// Select all transactions in the mempool that are valid and not already in the
	// partial proposal.
	for ; iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReachedBy(blockUsed) {
			break
		}

		tx := iterator.Tx()
		txInfo, err := l.GetTxInfo(tx)
		if err != nil {
			l.Logger.Info("failed to get hash of tx", "err", err)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// if the transaction is exceed the limit, we remove it from the lane mempool.
		if transactionLimit.IsExceededBy(txInfo.BlockSpace) {
			l.Logger.Info(
				"failed to select tx for lane; tx exceeds limit",
				"tx_hash", txInfo.Hash,
				"lane", l.Name,
			)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// Add the transaction to the proposal.
		if err := proposal.Add(txInfo); err != nil {
			l.Logger.Info(
				"failed to add tx to proposal",
				"lane", l.Name,
				"tx_hash", txInfo.Hash,
				"err", err,
			)

			break
		}

		// Update the total size and gas.
		blockUsed.IncreaseBy(txInfo.BlockSpace)
	}

	return
}

// GetTxInfo returns various information about the transaction that
// belongs to the lane including its priority, signer's, sequence number,
// size and more.
func (l *Lane) GetTxInfo(tx sdk.Tx) (TxWithInfo, error) {
	txBytes, err := l.TxEncoder(tx)
	if err != nil {
		return TxWithInfo{}, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// TODO: Add an adapter to lanes so that this can be flexible to support EVM, etc.
	gasTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return TxWithInfo{}, fmt.Errorf("failed to cast transaction to gas tx")
	}

	signers, err := l.SignerExtractor.GetSigners(tx)
	if err != nil {
		return TxWithInfo{}, err
	}

	BlockSpace := NewBlockSpace(int64(len(txBytes)), gasTx.GetGas())

	return TxWithInfo{
		Hash:       strings.ToUpper(hex.EncodeToString(comettypes.Tx(txBytes).Hash())),
		BlockSpace: BlockSpace,
		TxBytes:    txBytes,
		Signers:    signers,
	}, nil
}

// SetIsBlocked sets the isBlocked flag to the given value.
func (l *Lane) SetIsBlocked(isBlocked bool) {
	l.isBlocked = isBlocked
}
