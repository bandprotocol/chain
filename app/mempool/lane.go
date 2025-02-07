package mempool

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Lane defines a logical grouping of transactions within the mempool.
type Lane struct {
	// Name is a friendly identifier for this lane (e.g. "bankSend", "delegate").
	Name string

	// Filter determines if a given transaction belongs in this lane.
	Filter func(sdk.Tx) bool

	// Percentage is the fraction (0-100) of the block's gas budget allocated to this lane.
	Percentage uint64

	// EnforceOneTxPerSigner indicates that each signer may only have one tx
	// included in this lane per proposal.
	EnforceOneTxPerSigner bool

	// signersUsed tracks which signers have already added a transaction in this lane
	// for the current proposal. (Only used if EnforceOneTxPerSigner == true.)
	signersUsed map[string]bool

	// txs holds the set of transactions that matched Filter.
	txs []TxWithInfo
}

// NewLane is a simple constructor for a lane.
func NewLane(name string, filter func(sdk.Tx) bool, percentage uint64, enforceOneTxPerSigner bool) *Lane {
	return &Lane{
		Name:                  name,
		Filter:                filter,
		Percentage:            percentage,
		EnforceOneTxPerSigner: enforceOneTxPerSigner,
		signersUsed:           make(map[string]bool),
		txs:                   []TxWithInfo{},
	}
}

// AddTx appends a transaction to this lane's slice.
func (l *Lane) AddTx(tx TxWithInfo) {
	l.txs = append(l.txs, tx)
}

// RemoveTx removes a transaction from this lane by matching TxWithInfo.Hash.
func (l *Lane) RemoveTx(txInfo TxWithInfo) {
	newTxs := make([]TxWithInfo, 0, len(l.txs))
	for _, t := range l.txs {
		if t.Hash != txInfo.Hash {
			newTxs = append(newTxs, t)
		}
	}
	l.txs = newTxs
}

// GetTxs returns the slice of lane transactions.
func (l *Lane) GetTxs() []TxWithInfo {
	return l.txs
}

// SetTxs overwrites the lane's transactions with the new slice.
func (l *Lane) SetTxs(newTxs []TxWithInfo) {
	l.txs = newTxs
}

// ResetState resets the lane’s state for a new proposal,
// clearing the signersUsed map.
func (l *Lane) ResetState() {
	l.signersUsed = make(map[string]bool)
}

// FillProposal attempts to add lane transactions to the proposal,
// respecting the laneGasLimit and the “one tx per signer” rule (if enforced).
// It returns the leftover (unconsumed) gas for the lane.
func (l *Lane) FillProposal(proposal *Proposal, laneGasLimit uint64) uint64 {
	for _, txInfo := range l.txs {
		// If the next tx doesn't fit the lane's gas budget, skip it (or break).
		if txInfo.GasLimit > laneGasLimit {
			continue
		}

		// If we enforce "one tx per signer," check if any signer has already used this lane.
		if l.EnforceOneTxPerSigner {
			skip := false
			for _, signer := range txInfo.Signers {
				signerAddr := signer.Signer.String()
				if l.signersUsed[signerAddr] {
					// This signer has already used the lane for a prior tx.
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}

		// Attempt to add the transaction to the proposal.
		if err := proposal.Add(txInfo); err != nil {
			// If we fail (e.g. block size or gas limit exceeded), skip or break based on your policy.
			continue
		}

		// The transaction fits the lane's gas limit and the proposal’s constraints.
		laneGasLimit -= txInfo.GasLimit

		// Mark signers as used if one-tx-per-signer is enforced.
		if l.EnforceOneTxPerSigner {
			for _, signer := range txInfo.Signers {
				signerAddr := signer.Signer.String()
				l.signersUsed[signerAddr] = true
			}
		}
	}

	return laneGasLimit
}
