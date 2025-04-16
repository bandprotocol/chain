package yoda

import (
	"context"
	"fmt"
	"strings"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func signAndBroadcast(
	c *Context, key *keyring.Record, msgs []sdk.Msg, gasLimit uint64, memo string,
) (*sdk.TxResponse, error) {
	clientCtx := client.Context{
		Client:            c.client,
		Codec:             c.encodingConfig.Codec,
		TxConfig:          c.encodingConfig.TxConfig,
		BroadcastMode:     "sync",
		InterfaceRegistry: c.encodingConfig.InterfaceRegistry,
	}
	acc, err := queryAccount(clientCtx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get account: %w", err)
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(clientCtx.TxConfig).
		WithGas(gasLimit).WithGasAdjustment(1).
		WithChainID(cfg.ChainID).
		WithMemo(memo).
		WithGasPrices(c.gasPrices).
		WithKeybase(kb).
		WithAccountRetriever(clientCtx.AccountRetriever)

	address, err := key.GetAddress()
	if err != nil {
		return nil, err
	}

	execMsg := authz.NewMsgExec(address, msgs)

	txb, err := txf.BuildUnsignedTx(&execMsg)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(context.Background(), txf, key.Name, txb, true)
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

func SubmitReport(c *Context, l *Logger, keyIndex int64, reports []ReportMsgWithKey) {
	// Return key and update pending metric when done with SubmitReport whether successfully or not.
	defer func() {
		c.freeKeys <- keyIndex
	}()
	defer c.updatePendingGauge(int64(-len(reports)))

	// Summarize execute version
	versionMap := make(map[string]bool)
	msgs := make([]sdk.Msg, len(reports))
	ids := make([]types.RequestID, len(reports))
	feeEstimations := make([]FeeEstimationData, len(reports))

	for i, report := range reports {
		if err := report.msg.ValidateBasic(); err != nil {
			l.Error(":exploding_head: Failed to validate basic with error: %s", c, err.Error())
			return
		}
		msgs[i] = report.msg
		ids[i] = report.msg.RequestID
		feeEstimations[i] = report.feeEstimationData
		for _, exec := range report.execVersion {
			versionMap[exec] = true
		}
	}
	l = l.With("rids", ids)

	versions := make([]string, 0, len(versionMap))
	for exec := range versionMap {
		versions = append(versions, exec)
	}
	memo := fmt.Sprintf("yoda:%s/exec:%s", version.Version, strings.Join(versions, ","))
	key := c.keys[keyIndex]

	clientCtx := client.Context{
		Client:            c.client,
		TxConfig:          c.encodingConfig.TxConfig,
		InterfaceRegistry: c.encodingConfig.InterfaceRegistry,
	}

	gasLimit := estimateGas(c, l, msgs, feeEstimations)
	// We want to resend transaction only if tx returns Out of gas error.
	for sendAttempt := uint64(1); sendAttempt <= c.maxTry; sendAttempt++ {
		var txHash string
		l.Info(":e-mail: Sending report transaction attempt: (%d/%d)", sendAttempt, c.maxTry)
		for broadcastTry := uint64(1); broadcastTry <= c.maxTry; broadcastTry++ {
			l.Info(":writing_hand: Try to sign and broadcast report transaction(%d/%d)", broadcastTry, c.maxTry)
			txResp, err := signAndBroadcast(c, key, msgs, gasLimit, memo)
			if err != nil {
				// Use info level because this error can happen and retry process can solve this error.
				l.Info(":warning: %s", err.Error())
				time.Sleep(c.rpcPollInterval)
				continue
			}

			// if the transaction is out of gas, increase the gas adjustment
			if txResp.Codespace == sdkerrors.RootCodespace && txResp.Code == sdkerrors.ErrOutOfGas.ABCICode() {
				gasLimit = gasLimit * 110 / 100
				l.Info(
					":fuel_pump: Tx(%s) is out of gas and will be rebroadcasted with %d gas",
					txResp.TxHash,
					gasLimit,
				)
				continue
			} else if txResp.Code != 0 {
				l.Error(":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s", c, txResp.Code, txResp.RawLog, txResp.TxHash)
				continue
			}
			// Transaction passed CheckTx process and wait to include in block.
			txHash = txResp.TxHash
			break
		}
		if txHash == "" {
			l.Error(":exploding_head: Cannot try to broadcast more than %d try", c, c.maxTry)
			return
		}
		txFound := false
	FindTx:
		for start := time.Now(); time.Since(start) < c.broadcastTimeout; {
			time.Sleep(c.rpcPollInterval)
			txResp, err := authtx.QueryTx(clientCtx, txHash)
			if err != nil {
				l.Debug(":warning: Failed to query tx with error: %s", err.Error())
				continue
			}

			if txResp.Code == 0 {
				l.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", txHash)
				c.updateSubmittedCount(int64(len(reports)))
				return
			}
			if txResp.Codespace == sdkerrors.RootCodespace &&
				txResp.Code == sdkerrors.ErrOutOfGas.ABCICode() {
				// Increase gas limit and try to broadcast again
				gasLimit = gasLimit * 110 / 100
				l.Info(":fuel_pump: Tx(%s) is out of gas and will be rebroadcasted with %d gas", txHash, gasLimit)
				txFound = true
				break FindTx
			} else {
				l.Error(":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s", c, txResp.Code, txResp.RawLog, txResp.TxHash)
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
	l.Error(":anxious_face_with_sweat: Cannot send reports with adjusted gas: %d", c, gasLimit)
}

// GetExecutable fetches data source executable using the provided client.
func GetExecutable(c *Context, l *Logger, hash string) ([]byte, error) {
	resValue, err := c.fileCache.GetFile(hash)
	if err != nil {
		l.Debug(":magnifying_glass_tilted_left: Fetching data source hash: %s from bandchain querier", hash)
		bz := c.encodingConfig.Codec.MustMarshal(&types.QueryDataRequest{
			DataHash: hash,
		})
		res, err := abciQuery(c, l, "/band.oracle.v1.Query/Data", bz)
		if err != nil {
			l.Error(":exploding_head: Failed to get data source with error: %s", c, err.Error())
			return nil, err
		}
		var dr types.QueryDataResponse
		err = c.encodingConfig.Codec.Unmarshal(res.Response.GetValue(), &dr)
		if err != nil {
			l.Error(":exploding_head: Failed to unmarshal data source with error: %s", c, err.Error())
			return nil, err
		}
		resValue = dr.Data
		c.fileCache.AddFile(resValue)
	} else {
		l.Debug(":card_file_box: Found data source hash: %s in cache file", hash)
	}

	l.Debug(":balloon: Received data source hash: %s content: %q", hash, resValue[:32])
	return resValue, nil
}

// GetDataSourceHash fetches data source hash by id
func GetDataSourceHash(c *Context, l *Logger, id types.DataSourceID) (string, error) {
	res, err := abciQuery(c, l, fmt.Sprintf("/store/%s/key", types.StoreKey), types.DataSourceStoreKey(id))
	if err != nil {
		l.Error(":skull: Failed to get data source with error: %s", c, err.Error())
		return "", err
	}

	var d types.DataSource
	c.encodingConfig.Codec.MustUnmarshal(res.Response.Value, &d)

	return d.Filename, nil
}

// GetRequest fetches request by id
func GetRequest(c *Context, l *Logger, id types.RequestID) (types.Request, error) {
	res, err := abciQuery(c, l, fmt.Sprintf("/store/%s/key", types.StoreKey), types.RequestStoreKey(id))
	if err != nil {
		l.Error(":skull: Failed to get request with error: %s", c, err.Error())
		return types.Request{}, err
	}

	var r types.Request
	c.encodingConfig.Codec.MustUnmarshal(res.Response.Value, &r)

	return r, nil
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
