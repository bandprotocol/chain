package mempool

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"

	comettypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

// Lane defines a logical grouping of transactions within the mempool.
type Lane struct {
	Logger          log.Logger
	TxEncoder       sdk.TxEncoder
	SignerExtractor signerextraction.Adapter
	// Name is a identifier for this lane (e.g. "bankSend", "delegate").
	Name string

	// Filter determines if a given transaction belongs in this lane.
	Match func(sdk.Tx) bool

	MaxTransactionSpace math.LegacyDec
	MaxLaneSpace        math.LegacyDec

	// laneMempool is the mempool that is responsible for storing transactions
	// that are waiting to be processed.
	laneMempool sdkmempool.Mempool
}

// NewLane is a simple constructor for a lane.
func NewLane(
	logger log.Logger,
	txEncoder sdk.TxEncoder,
	signerExtractor signerextraction.Adapter,
	name string,
	matchFn func(sdk.Tx) bool,
	maxTransactionSpace math.LegacyDec,
	maxLaneSpace math.LegacyDec,
	laneMempool sdkmempool.Mempool,
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
	}
}

func (l *Lane) Insert(ctx context.Context, tx sdk.Tx) error {
	return l.laneMempool.Insert(ctx, tx)
}

func (l *Lane) CountTx() int {
	return l.laneMempool.CountTx()
}

func (l *Lane) Remove(tx sdk.Tx) error {
	return l.laneMempool.Remove(tx)
}

// FillProposal fills the proposal with transactions from the lane mempool.
// It returns the total size and gas of the transactions added to the proposal.
// If customLaneLimit is provided, it will be used instead of the lane's limit.
func (l *Lane) FillProposal(
	ctx sdk.Context,
	proposal *Proposal,
) (sizeUsed int64, gasUsed uint64, iterator sdkmempool.Iterator, txsToRemove []sdk.Tx) {
	var (
		transactionLimit BlockSpace
		laneLimit        BlockSpace
	)
	// Get the transaction and lane limit for the lane.
	transactionLimit = proposal.GetLimit(l.MaxTransactionSpace)
	laneLimit = proposal.GetLimit(l.MaxLaneSpace)

	// Select all transactions in the mempool that are valid and not already in the
	// partial proposal.
	for iterator = l.laneMempool.Select(ctx, nil); iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReached(sizeUsed, gasUsed) {

			break
		}

		tx := iterator.Tx()
		txInfo, err := l.GetTxInfo(ctx, tx)
		if err != nil {
			l.Logger.Info("failed to get hash of tx", "err", err)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// if the transaction is exceed the limit, we remove it from the lane mempool.
		if transactionLimit.IsExceeded(txInfo.Size, txInfo.GasLimit) {
			l.Logger.Info(
				"failed to select tx for lane; tx exceeds limit",
				"tx_hash", txInfo.Hash,
				"lane", l.Name,
			)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// Verify the transaction.
		if err = l.VerifyTx(ctx, tx, false); err != nil {
			l.Logger.Info(
				"failed to verify tx",
				"tx_hash", txInfo.Hash,
				"err", err,
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

		// Update the total size and gas.
		sizeUsed += txInfo.Size
		gasUsed += txInfo.GasLimit
	}

	return
}

func (l *Lane) FillProposalBy(
	ctx sdk.Context,
	proposal *Proposal,
	iterator sdkmempool.Iterator,
	laneLimit BlockSpace,
) (sizeUsed int64, gasUsed uint64, txsToRemove []sdk.Tx) {
	// get the transaction limit for the lane.
	transactionLimit := proposal.GetLimit(l.MaxTransactionSpace)

	// Select all transactions in the mempool that are valid and not already in the
	// partial proposal.
	for ; iterator != nil; iterator = iterator.Next() {
		// If the total size used or total gas used exceeds the limit, we break and do not attempt to include more txs.
		// We can tolerate a few bytes/gas over the limit, since we limit the size of each transaction.
		if laneLimit.IsReached(sizeUsed, gasUsed) {
			break
		}

		tx := iterator.Tx()
		txInfo, err := l.GetTxInfo(ctx, tx)
		if err != nil {
			l.Logger.Info("failed to get hash of tx", "err", err)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// if the transaction is exceed the limit, we remove it from the lane mempool.
		if transactionLimit.IsExceeded(txInfo.Size, txInfo.GasLimit) {
			l.Logger.Info(
				"failed to select tx for lane; tx exceeds limit",
				"tx_hash", txInfo.Hash,
				"lane", l.Name,
			)

			txsToRemove = append(txsToRemove, tx)
			continue
		}

		// Verify the transaction.
		if err = l.VerifyTx(ctx, tx, false); err != nil {
			l.Logger.Info(
				"failed to verify tx",
				"tx_hash", txInfo.Hash,
				"err", err,
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
		sizeUsed += txInfo.Size
		gasUsed += txInfo.GasLimit
	}

	return
}

// GetTxInfo returns various information about the transaction that
// belongs to the lane including its priority, signer's, sequence number,
// size and more.
func (l *Lane) GetTxInfo(ctx sdk.Context, tx sdk.Tx) (TxWithInfo, error) {
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

	return TxWithInfo{
		Hash:     strings.ToUpper(hex.EncodeToString(comettypes.Tx(txBytes).Hash())),
		Size:     int64(len(txBytes)),
		GasLimit: gasTx.GetGas(),
		TxBytes:  txBytes,
		Signers:  signers,
	}, nil
}

// TODO: Add a method to verify the transaction.
// VerifyTx verifies that the transaction is valid respecting to the msg server
func (l *Lane) VerifyTx(ctx sdk.Context, tx sdk.Tx, simulate bool) error {
	return nil
}
