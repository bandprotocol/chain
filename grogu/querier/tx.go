package querier

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type TxQuerier struct {
	contexts []client.Context
}

func NewTxQuerier(clients []client.Context) *TxQuerier {
	return &TxQuerier{clients}
}

func (q *TxQuerier) QueryTx(hash string) (*types.TxResponse, error) {
	resultCh := make(chan *types.TxResponse, len(q.contexts))
	failureCh := make(chan error, len(q.contexts))

	for _, ctx := range q.contexts {
		go func(ctx client.Context) {
			resp, err := tx.QueryTx(ctx, hash)
			if err != nil {
				failureCh <- err
				return
			}

			resultCh <- resp
		}(ctx)
	}

	var err error
	for range q.contexts {
		select {
		case res := <-resultCh:
			return res, nil
		case err = <-failureCh:
			continue
		}
	}

	return nil, err
}
