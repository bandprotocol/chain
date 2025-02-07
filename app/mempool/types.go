package mempool

import (
	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
)

// TxWithInfo holds metadata required for a transaction to be included in a proposal.
type TxWithInfo struct {
	// Hash is the hex-encoded hash of the transaction.
	Hash string
	// Size is the size of the transaction in bytes.
	Size int64
	// GasLimit is the gas limit of the transaction.
	GasLimit uint64
	// TxBytes is the raw transaction bytes.
	TxBytes []byte
	// Signers defines the signers of a transaction.
	Signers []signerextraction.SignerData
}
