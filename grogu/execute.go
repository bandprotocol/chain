package grogu

import (
	"context"
	"fmt"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	band "github.com/bandprotocol/chain/v2/app"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
)

var (
	// Proto codec for encoding/decoding proto message
	cdc = band.MakeEncodingConfig().Marshaler
)

func signAndBroadcast(
	c *Context, key *keyring.Record, msgs []sdk.Msg, l *Logger,
) (string, error) {
	l.Info("exp 1")
	clientCtx := client.Context{
		Client:            c.client,
		Codec:             cdc,
		TxConfig:          band.MakeEncodingConfig().TxConfig,
		BroadcastMode:     flags.BroadcastSync,
		InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
	}
	l.Info("exp 2")
	acc, err := queryAccount(clientCtx, key)
	if err != nil {
		return "", fmt.Errorf("unable to get account: %w", err)
	}
	l.Info("exp 3")
	l.Info(c.gasPrices)

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(band.MakeEncodingConfig().TxConfig).
		WithSimulateAndExecute(true).
		WithGasAdjustment(2).
		WithChainID(cfg.ChainID).
		WithGasPrices(c.gasPrices).
		WithKeybase(kb).
		WithAccountRetriever(clientCtx.AccountRetriever)
	l.Info("exp 4")

	fmt.Printf("num: %+v\n", acc.GetAccountNumber())
	fmt.Printf("seq: %+v\n", acc.GetSequence())
	fmt.Printf(": %+v\n", acc)
	address, err := key.GetAddress()
	if err != nil {
		return "", err
	}
	l.Info("exp 5")

	fmt.Printf("num: %+v\n", acc.GetAccountNumber())
	fmt.Printf("seq: %+v\n", acc.GetSequence())
	fmt.Printf("address: %+v\n", address.String())
	fmt.Printf("msgs: %+v\n", msgs)

	execMsg := authz.NewMsgExec(address, msgs)
	l.Info("exp 6")

	_, adjusted, err := tx.CalculateGas(clientCtx, txf, &execMsg)
	if err != nil {
		return "", err
	}

	// Set the gas amount on the transaction factory
	txf = txf.WithGas(adjusted)

	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return "", err
	}
	l.Info("exp 7")

	err = tx.Sign(txf, key.Name, txb, true)
	if err != nil {
		return "", err
	}
	l.Info("exp 8")

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return "", err
	}
	l.Info("exp 9")

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", err
	}
	l.Info("exp 10")
	// out, err := txBldr.WithKeybase(keybase).BuildAndSign(key.GetName(), ckeys.DefaultKeyPass, msgs)
	// if err != nil {
	// 	return "", fmt.Errorf("Failed to build tx with error: %s", err.Error())
	// }
	return res.TxHash, nil
}

func queryAccount(clientCtx client.Context, key *keyring.Record) (client.Account, error) {
	accountRetriever := authtypes.AccountRetriever{}

	address, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	acc, err := accountRetriever.GetAccount(clientCtx, address)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func SubmitPrices(c *Context, l *Logger, keyIndex int64, prices []feedstypes.SubmitPrice) {
	l.Info("inside SubmitPrices")
	// Return key and update pending metric when done with SubmitReport whether successfully or not.
	defer func() {
		c.freeKeys <- keyIndex
	}()

	msg := feedstypes.MsgSubmitPrices{
		Validator: c.validator.String(),
		Timestamp: time.Now().Unix(),
		Prices:    prices,
	}

	msgs := []sdk.Msg{&msg}
	l.Info("before key")
	key := c.keys[keyIndex]
	l.Info("keys", key)
	hash, err := signAndBroadcast(c, key, msgs, l)
	if err != nil {
		l.Info(":warning:  err:%s", err.Error())
	}
	l.Info("hash %s", hash)
}

// abciQuery will try to query data from BandChain node maxTry time before give up and return error
func abciQuery(c *Context, l *Logger, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	var lastErr error
	for try := 0; try < int(c.maxTry); try++ {
		res, err := c.client.ABCIQuery(context.Background(), path, data)
		if err != nil {
			l.Debug(":skull: Failed to query on %s request with error: %s", path, err.Error())
			lastErr = err
			time.Sleep(c.rpcPollInterval)
			continue
		}
		return res, nil
	}
	return nil, lastErr
}
