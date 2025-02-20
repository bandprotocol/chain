package mempool

import (
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
)

// TxWithInfo holds metadata required for a transaction to be included in a proposal.
type TxWithInfo struct {
	// Hash is the hex-encoded hash of the transaction.
	Hash string
	// BlockSpace is the block space used by the transaction.
	BlockSpace BlockSpace
	// TxBytes is the raw transaction bytes.
	TxBytes []byte
	// Signers defines the signers of a transaction.
	Signers []sdkmempool.SignerData
}
