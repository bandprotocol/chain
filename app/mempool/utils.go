package mempool

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetDecodedTxs returns the decoded transactions from the given bytes.
func GetDecodedTxs(txDecoder sdk.TxDecoder, txs [][]byte) ([]sdk.Tx, error) {
	var decodedTxs []sdk.Tx
	for _, txBz := range txs {
		tx, err := txDecoder(txBz)
		if err != nil {
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}

		decodedTxs = append(decodedTxs, tx)
	}

	return decodedTxs, nil
}
