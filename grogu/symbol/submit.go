package symbol

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	band "github.com/bandprotocol/chain/v2/app"
	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func StartSubmitPrices(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		SubmitPrices(c, l)
	}
}

func SubmitPrices(c *grogucontext.Context, l *grogucontext.Logger) {
	// Return key and update pending metric when done with SubmitReport whether successfully or not.
	keyIndex := <-c.FreeKeys
	defer func() {
		c.FreeKeys <- keyIndex
	}()

	prices := <-c.PendingPrices

GetAllPrices:
	for {
		select {
		case nextPrices := <-c.PendingPrices:
			prices = append(prices, nextPrices...)
		default:
			break GetAllPrices
		}
	}

	defer func() {
		for _, price := range prices {
			c.InProgressSymbols.Delete(price.Symbol)
		}
	}()

	msg := types.MsgSubmitPrices{
		Validator: c.Validator.String(),
		Timestamp: time.Now().Unix(),
		Prices:    prices,
	}

	msgs := []sdk.Msg{&msg}
	key := c.Keys[keyIndex]

	clientCtx := client.Context{
		Client:            c.Client,
		TxConfig:          band.MakeEncodingConfig().TxConfig,
		InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
	}

	gasAdjustment := float64(2.0)

	for sendAttempt := uint64(1); sendAttempt <= c.MaxTry; sendAttempt++ {
		var txHash string
		l.Info(":e-mail: Sending report transaction attempt: (%d/%d)", sendAttempt, c.MaxTry)
		for broadcastTry := uint64(1); broadcastTry <= c.MaxTry; broadcastTry++ {
			l.Info(":writing_hand: Try to sign and broadcast report transaction(%d/%d)", broadcastTry, c.MaxTry)
			res, err := signAndBroadcast(c, key, msgs, gasAdjustment)
			if err != nil {
				// Use info level because this error can happen and retry process can solve this error.
				l.Info(":warning: %s", err.Error())
				time.Sleep(c.RPCPollInterval)
				continue
			}
			if res.Codespace == sdkerrors.RootCodespace && res.Code == sdkerrors.ErrOutOfGas.ABCICode() {
				gasAdjustment += 0.1
				l.Info(
					":fuel_pump: Tx(%s) is out of gas and will be rebroadcasted with gas adjustment(%f)",
					txHash,
					gasAdjustment,
				)
				continue
			}
			// Transaction passed CheckTx process and wait to include in block.
			txHash = res.TxHash
			break
		}
		if txHash == "" {
			l.Error(":exploding_head: Cannot try to broadcast more than %d try", c, c.MaxTry)
			return
		}
		txFound := false
	FindTx:
		for start := time.Now(); time.Since(start) < c.BroadcastTimeout; {
			time.Sleep(c.RPCPollInterval)
			txRes, err := authtx.QueryTx(clientCtx, txHash)
			if err != nil {
				l.Debug(":warning: Failed to query tx with error: %s", err.Error())
				continue
			}

			if txRes.Code == 0 {
				l.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", txHash)
				return
			}
			if txRes.Codespace == sdkerrors.RootCodespace &&
				txRes.Code == sdkerrors.ErrOutOfGas.ABCICode() {
				// Increase gas adjustment and try to broadcast again
				gasAdjustment += 0.1
				l.Info(":fuel_pump: Tx(%s) is out of gas and will be rebroadcasted with gas adjustment(%f)", txHash, gasAdjustment)
				txFound = true
				break FindTx
			} else {
				l.Error(":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s", c, txRes.Code, txRes.RawLog, txRes.TxHash)
				return
			}
		}
		if !txFound {
			l.Error(
				":question_mark: Cannot get transaction response from hash: %s transaction might be included in the next few blocks or check your node's health.",
				c,
				txHash,
			)
			return
		}
	}
	l.Error(":anxious_face_with_sweat: Cannot send price with adjusted gas: %d", c, gasAdjustment)
}

func signAndBroadcast(
	c *grogucontext.Context, key *keyring.Record, msgs []sdk.Msg, gasAdjustment float64,
) (*sdk.TxResponse, error) {
	clientCtx := client.Context{
		Client:            c.Client,
		Codec:             grogucontext.Cdc,
		TxConfig:          band.MakeEncodingConfig().TxConfig,
		BroadcastMode:     flags.BroadcastSync,
		InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
	}
	acc, err := queryAccount(clientCtx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(band.MakeEncodingConfig().TxConfig).
		WithSimulateAndExecute(true).
		WithGasAdjustment(gasAdjustment).
		WithChainID(grogucontext.Cfg.ChainID).
		WithGasPrices(c.GasPrices).
		WithKeybase(grogucontext.Kb).
		WithFromName(key.Name).
		WithAccountRetriever(clientCtx.AccountRetriever)

	address, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	execMsg := authz.NewMsgExec(address, msgs)

	_, adjusted, err := tx.CalculateGas(clientCtx, txf, &execMsg)
	if err != nil {
		return nil, err
	}

	// Set the gas amount on the transaction factory
	txf = txf.WithGas(adjusted)

	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(txf, key.Name, txb, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
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
