package submitter

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	band "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/grogu/querier"
)

func getTxResponse(
	txQuerier *querier.TxQuerier,
	txHash string,
	timeout time.Duration,
	pollInterval time.Duration,
) (*sdk.TxResponse, error) {
	var resp *sdk.TxResponse
	var err error

	for start := time.Now(); time.Since(start) < timeout; {
		time.Sleep(pollInterval)
		resp, err = txQuerier.QueryTx(txHash)
		if err != nil {
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("timeout exceeded with error: %v", err)
}

func broadcastMsg(
	ctx client.Context,
	account client.Account,
	key *keyring.Record,
	msgs []sdk.Msg,
	gasPrice string,
	gasAdjustment float64,
	memo string,
) (*sdk.TxResponse, error) {
	addr, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	txf := tx.Factory{}.
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence()).
		WithTxConfig(ctx.TxConfig).
		WithSimulateAndExecute(true).
		WithGasAdjustment(gasAdjustment).
		WithChainID(ctx.ChainID).
		WithMemo(memo).
		WithGasPrices(gasPrice).
		WithKeybase(ctx.Keyring).
		WithFromName(key.Name).
		WithAccountRetriever(ctx.AccountRetriever)

	execMsg := authz.NewMsgExec(addr, msgs)

	_, adjusted, err := tx.CalculateGas(ctx, txf, &execMsg)
	if err != nil {
		return nil, err
	}

	txf = txf.WithGas(adjusted)
	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(txf, key.Name, txb, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := ctx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	return ctx.BroadcastTx(txBytes)
}

func broadcastMsgWithMultipleContext(
	contexts []client.Context,
	account client.Account,
	key *keyring.Record,
	msgs []sdk.Msg,
	gasPrice string,
	gasAdjustment float64,
	memo string,
) (*sdk.TxResponse, error) {
	if len(contexts) == 0 {
		return nil, fmt.Errorf("no context provided")
	}

	resultsCh := make(chan *sdk.TxResponse, len(contexts))
	failureCh := make(chan error, len(contexts))
	for _, ctx := range contexts {
		go func(ctx client.Context) {
			res, err := broadcastMsg(ctx, account, key, msgs, gasPrice, gasAdjustment, memo)
			if err != nil {
				failureCh <- err
				return
			}
			resultsCh <- res
		}(ctx)
	}

	var res *sdk.TxResponse
	var err error
	for range contexts {
		select {
		case currentResult := <-resultsCh:
			if currentResult.Code == 0 {
				return currentResult, nil
			}

			res = currentResult
		case err = <-failureCh:
			continue
		}
	}

	if res != nil {
		return res, nil
	}

	return nil, err
}

func unpackAccount(account *client.Account, resp *auth.QueryAccountResponse) error {
	registry := band.MakeEncodingConfig().InterfaceRegistry
	err := registry.UnpackAny(resp.Account, account)
	if err != nil {
		return fmt.Errorf("failed to unpack account with error: %v", err)
	}

	return nil
}
