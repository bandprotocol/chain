package band

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	comettypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ sdkmempool.Mempool = (*BandMempool)(nil)

// BandMempool defines the Band mempool implementation.
type BandMempool struct {
	txEncoder sdk.TxEncoder
	// bankSendTxs contains the submit price transactions.
	bankSendTxs []TxWithInfo

	// delegateTxs contains the delegate transactions.
	delegateTxs []TxWithInfo

	// otherTxs contains the other transactions.
	otherTxs []TxWithInfo
}

// NewBandMempool returns a new BandMempool.
func NewBandMempool(txEncoder sdk.TxEncoder) *BandMempool {
	return &BandMempool{
		txEncoder:   txEncoder,
		bankSendTxs: []TxWithInfo{},
		delegateTxs: []TxWithInfo{},
		otherTxs:    []TxWithInfo{},
	}
}

// Insert inserts a transaction into the mempool.
func (m *BandMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	txInfo, err := m.GetTxInfo(tx)
	if err != nil {
		fmt.Println("failed to get hash of tx", "err", err)

		return err
	}

	// if tx is a bank send transaction
	if isBankSendTx(tx) {
		m.bankSendTxs = append(m.bankSendTxs, txInfo)
		return nil
	}

	// if tx is a delegate transaction
	if isDelegateTx(tx) {
		m.delegateTxs = append(m.delegateTxs, txInfo)
		return nil
	}

	// if tx is any other transaction
	m.otherTxs = append(m.otherTxs, txInfo)
	return nil
}

// Select returns an Iterator over the app-side mempool.
func (m *BandMempool) Select(ctx context.Context, txs [][]byte) sdkmempool.Iterator {
	return nil
}

// CountTx returns the number of transactions currently in the mempool.
func (m *BandMempool) CountTx() int {
	return len(m.bankSendTxs) + len(m.delegateTxs) + len(m.otherTxs)
}

// Remove attempts to remove a transaction from the mempool.
func (m *BandMempool) Remove(tx sdk.Tx) error {
	fmt.Println("Removing transaction", tx, "from mempool")
	txInfo, err := m.GetTxInfo(tx)
	if err != nil {
		fmt.Println("failed to get hash of tx", "err", err)

		return err
	}
	// Remove from bankSendTxs
	m.bankSendTxs = removeTxFromSlice(m.bankSendTxs, txInfo)
	// Remove from delegateTxs
	m.delegateTxs = removeTxFromSlice(m.delegateTxs, txInfo)
	// Remove from otherTxs
	m.otherTxs = removeTxFromSlice(m.otherTxs, txInfo)
	return nil
}

// PrepareBandProposal fills the proposal with transactions from each lane.
func (m *BandMempool) PrepareBandProposal(ctx sdk.Context, proposal Proposal) (Proposal, error) {
	fmt.Println("number of txs in bankSendTxs", len(m.bankSendTxs))
	fmt.Println("number of txs in delegateTxs", len(m.delegateTxs))
	fmt.Println("number of txs in otherTxs", len(m.otherTxs))
	// Calculate the gas limits for each category.
	totalGasLimit := proposal.Info.MaxGasLimit
	bankSendGasLimit := (30 * totalGasLimit) / 100
	delegateGasLimit := (30 * totalGasLimit) / 100
	otherGasLimit := (40 * totalGasLimit) / 100

	// Fill the proposal with bankSendTxs (up to 30% of the gas limit).
	proposal, bankSendRemaining := m.fillProposalWithTxs(proposal, m.bankSendTxs, bankSendGasLimit)
	if bankSendRemaining > 0 {
		// If there is remaining gas after filling bankSendTxs, use it for delegateTxs.
		delegateGasLimit += bankSendRemaining
	}

	// Fill the proposal with delegateTxs (up to 30% of the gas limit).
	proposal, delegateRemaining := m.fillProposalWithTxs(proposal, m.delegateTxs, delegateGasLimit)
	if delegateRemaining > 0 {
		// If there is remaining gas after filling delegateTxs, use it for otherTxs.
		otherGasLimit += delegateRemaining
	}

	// Fill the proposal with otherTxs (up to 40% of the gas limit).
	proposal, otherRemaining := m.fillProposalWithTxs(proposal, m.otherTxs, otherGasLimit)
	if otherRemaining > 0 {
		// If there is still remaining gas, fill it in the order: bankSendTxs -> delegateTxs -> otherTxs.
		proposal, otherRemaining = m.fillProposalWithTxs(proposal, m.bankSendTxs, otherRemaining)
		proposal, otherRemaining = m.fillProposalWithTxs(proposal, m.delegateTxs, otherRemaining)
		proposal, _ = m.fillProposalWithTxs(proposal, m.otherTxs, otherRemaining)
	}

	return proposal, nil
}

// fillProposalWithTxs fills the proposal with transactions from a given slice until the gas limit is reached.
// It returns the updated proposal and the remaining gas limit.
func (m *BandMempool) fillProposalWithTxs(
	proposal Proposal,
	txInfos []TxWithInfo,
	gasLimit uint64,
) (Proposal, uint64) {
	for _, txInfo := range txInfos {
		if txInfo.GasLimit > gasLimit {
			// TODO: consider continuing to the next transaction if the gas limit is exceeded.
			break
		}

		if err := proposal.Add(txInfo); err != nil {
			// TODO: consider break in case of full proposal.
			continue
		}

		gasLimit -= txInfo.GasLimit
	}

	return proposal, gasLimit
}

func (m *BandMempool) GetTxInfo(tx sdk.Tx) (TxWithInfo, error) {
	txBytes, err := m.txEncoder(tx)
	if err != nil {
		return TxWithInfo{}, fmt.Errorf("failed to encode transaction: %w", err)
	}

	// TODO: Add an adapter to lanes so that this can be flexible to support EVM, etc.
	gasTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return TxWithInfo{}, fmt.Errorf("failed to cast transaction to gas tx")
	}

	return TxWithInfo{
		Hash:     strings.ToUpper(hex.EncodeToString(comettypes.Tx(txBytes).Hash())),
		Size:     int64(len(txBytes)),
		GasLimit: gasTx.GetGas(),
		TxBytes:  txBytes,
	}, nil
}

// TxWithInfo contains the information required for a transaction to be
// included in a proposal.
type TxWithInfo struct {
	// Hash is the hex-encoded hash of the transaction.
	Hash string
	// Size is the size of the transaction in bytes.
	Size int64
	// GasLimit is the gas limit of the transaction.
	GasLimit uint64
	// TxBytes is the bytes of the transaction.
	TxBytes []byte
}

// removeTxFromSlice removes a transaction from a slice of transactions.
func removeTxFromSlice(txInfos []TxWithInfo, txInfo TxWithInfo) []TxWithInfo {
	for i, t := range txInfos {
		if t.Hash == txInfo.Hash {
			fmt.Println("****** Removing transaction", txInfo.Hash, "from slice")
			return append(txInfos[:i], txInfos[i+1:]...)
		}
	}
	return txInfos
}

// BandMempoolIterator is an iterator over the BandMempool.
type BandMempoolIterator struct {
	txs   []sdk.Tx
	index int
}

// Next returns the next iterator. If there are no more transactions, it returns nil.
func (it *BandMempoolIterator) Next() sdkmempool.Iterator {
	it.index++
	if it.index >= len(it.txs) {
		return nil
	}
	return it
}

// Tx returns the transaction at the current position of the iterator.
func (it *BandMempoolIterator) Tx() sdk.Tx {
	if it.index < len(it.txs) {
		return it.txs[it.index]
	}
	return nil
}

// isBankSendTx returns true if the transaction is a bank send transaction.
func isBankSendTx(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}

	for _, msg := range msgs {
		if _, ok := msg.(*banktypes.MsgSend); !ok {
			return false
		}
	}

	return true
}

// isDelegateTx returns true if the transaction is a delegate transaction.
func isDelegateTx(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}

	for _, msg := range msgs {
		if _, ok := msg.(*stakingtypes.MsgDelegate); !ok {
			return false
		}
	}

	return true
}
