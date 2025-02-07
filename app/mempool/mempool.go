package mempool

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"

	comettypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

// Mempool implements the sdkmempool.Mempool interface and uses Lanes internally.
type Mempool struct {
	txEncoder       sdk.TxEncoder
	signerExtractor signerextraction.Adapter
	lanes           []*Lane
}

// NewMempool returns a new mempool with the given lanes.
func NewMempool(
	txEncoder sdk.TxEncoder,
	signerExtractor signerextraction.Adapter,
	lanes []*Lane,
) *Mempool {
	return &Mempool{
		txEncoder:       txEncoder,
		signerExtractor: signerExtractor,
		lanes:           lanes,
	}
}

var _ sdkmempool.Mempool = (*Mempool)(nil)

// Insert inserts a transaction into the first matching lane.
func (m *Mempool) Insert(ctx context.Context, tx sdk.Tx) error {
	txInfo, err := m.GetTxInfo(tx)
	if err != nil {
		return fmt.Errorf("Insert: failed to get tx info: %w", err)
	}

	placed := false
	for _, lane := range m.lanes {
		if lane.Filter(tx) {
			lane.AddTx(txInfo)
			placed = true
			break
		}
	}
	if !placed {
		fmt.Printf("Insert: no lane matched for tx %s\n", txInfo.Hash)
	}
	return nil
}

// Select returns a Mempool iterator (currently nil).
func (m *Mempool) Select(ctx context.Context, txs [][]byte) sdkmempool.Iterator {
	return nil
}

// CountTx returns the total number of transactions across all lanes.
func (m *Mempool) CountTx() int {
	count := 0
	for _, lane := range m.lanes {
		count += len(lane.GetTxs())
	}
	return count
}

// Remove attempts to remove a transaction from whichever lane it is in.
func (m *Mempool) Remove(tx sdk.Tx) error {
	txInfo, err := m.GetTxInfo(tx)
	if err != nil {
		return fmt.Errorf("Remove: failed to get tx info: %w", err)
	}
	for _, lane := range m.lanes {
		lane.RemoveTx(txInfo)
	}
	return nil
}

// PrepareProposal divides the block gas limit among lanes (based on lane percentage),
// then calls each lane’s FillProposal method. If leftover gas is important to you,
// you can implement a second pass or distribute leftover to subsequent lanes, etc.
func (m *Mempool) PrepareProposal(ctx sdk.Context, proposal Proposal) (Proposal, error) {
	// Reset each lane’s state for the new proposal, clearing signersUsed, etc.
	for _, lane := range m.lanes {
		lane.ResetState()
	}

	totalGasLimit := proposal.Info.MaxGasLimit

	// 1) Compute each lane's gas budget from its percentage
	laneGasLimits := make([]uint64, len(m.lanes))
	for i, lane := range m.lanes {
		laneGasLimits[i] = (lane.Percentage * totalGasLimit) / 100
	}

	// 2) Fill the proposal lane by lane
	remainder := uint64(0)
	for i, lane := range m.lanes {
		remainder += lane.FillProposal(&proposal, laneGasLimits[i])
	}

	// 3) use remainder gas from first round to fill proposal further
	for _, lane := range m.lanes {
		remainder = lane.FillProposal(&proposal, remainder)
	}
	return proposal, nil
}

// GetTxInfo returns metadata (hash, size, gas, signers) for the given tx by encoding it.
func (m *Mempool) GetTxInfo(tx sdk.Tx) (TxWithInfo, error) {
	txBytes, err := m.txEncoder(tx)
	if err != nil {
		return TxWithInfo{}, fmt.Errorf("failed to encode transaction: %w", err)
	}

	gasTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return TxWithInfo{}, fmt.Errorf("failed to cast transaction to FeeTx")
	}

	signers, err := m.signerExtractor.GetSigners(tx)
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

// MempoolIterator is an example iterator (optional).
type MempoolIterator struct {
	txs   []sdk.Tx
	index int
}

func (it *MempoolIterator) Next() sdkmempool.Iterator {
	it.index++
	if it.index >= len(it.txs) {
		return nil
	}
	return it
}

func (it *MempoolIterator) Tx() sdk.Tx {
	if it.index < len(it.txs) {
		return it.txs[it.index]
	}
	return nil
}
