package yoda

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	odin "github.com/GeoDB-Limited/odin-core/app"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

var (
	// Use this as codec to legacy msg
	cdc = odin.MakeEncodingConfig().Amino
)

func signAndBroadcast(ctx *Context, key keyring.Info, msgs []sdk.Msg, gasLimit uint64, memo string) (string, error) {
	clientCtx := client.Context{
		Client:            ctx.client,
		TxConfig:          odin.MakeEncodingConfig().TxConfig,
		BroadcastMode:     flags.BroadcastAsync,
		InterfaceRegistry: odin.MakeEncodingConfig().InterfaceRegistry,
	}
	accountRetriever := authtypes.AccountRetriever{}
	acc, err := accountRetriever.GetAccount(clientCtx, key.GetAddress())
	if err != nil {
		return "", sdkerrors.Wrap(err, "failed to retrieve account")
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(odin.MakeEncodingConfig().TxConfig).
		WithGas(gasLimit).WithGasAdjustment(1).
		WithChainID(yoda.config.ChainID).
		WithMemo(memo).
		WithGasPrices(ctx.gasPrices).
		WithKeybase(yoda.keybase).
		WithAccountRetriever(clientCtx.AccountRetriever)

	txb, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return "", sdkerrors.Wrap(err, "failed to build unsigned tx")
	}

	err = tx.Sign(txf, key.GetName(), txb, true)
	if err != nil {
		return "", sdkerrors.Wrap(err, "failed to sign transaction")
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return "", sdkerrors.Wrap(err, "failed to encode transaction")
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", sdkerrors.Wrap(err, "failed to broadcast transaction")
	}

	return res.TxHash, nil
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
	// cliCtx := sdkCtx.CLIContext{Client: c.client, TrustNode: true, Codec: cdc}
	clientCtx := client.Context{Client: c.client, TxConfig: odin.MakeEncodingConfig().TxConfig}
	gasLimit := estimateGas(c, msgs, feeEstimations)
	// We want to resend transaction only if tx returns Out of gas error.
	for sendAttempt := uint64(1); sendAttempt <= c.maxTry; sendAttempt++ {
		var txHash string
		l.Info(":e-mail: Sending report transaction attempt: (%d/%d)", sendAttempt, c.maxTry)
		for broadcastTry := uint64(1); broadcastTry <= c.maxTry; broadcastTry++ {
			l.Info(
				":writing_hand: Try to sign and broadcast report transaction(%d/%d) with gas limit: %d",
				broadcastTry,
				c.maxTry,
				gasLimit,
			)
			hash, err := signAndBroadcast(c, key, msgs, gasLimit, memo)
			if err != nil {
				// Use info level because this error can happen and retry process can solve this error.
				l.Info(":warning: %s", err.Error())
				time.Sleep(c.rpcPollInterval)
				continue
			}
			// Transaction passed CheckTx process and wait to include in block.
			txHash = hash
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
			txRes, err := authclient.QueryTx(clientCtx, txHash)
			if err != nil {
				l.Debug(":warning: Failed to query tx with error: %s", err.Error())
				continue
			}
			switch txRes.Code {
			case 0:
				l.Info(":smiling_face_with_sunglasses: Successfully broadcast tx with hash: %s", txHash)
				c.updateSubmittedCount(int64(len(reports)))
				return
			case sdkerrors.ErrOutOfGas.ABCICode():
				// Increase gas limit and try to broadcast again
				gasLimit = gasLimit * 110 / 100
				l.Info(
					":fuel_pump: Tx(%s) is out of gas and will be rebroadcasted with %d gas",
					txHash,
					gasLimit,
				)
				txFound = true
				break FindTx
			default:
				l.Error(
					":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s",
					c,
					txRes.Code,
					txRes.RawLog,
					txRes.TxHash,
				)
				return
			}
		}
		if !txFound {
			l.Error(":question_mark: Cannot get transaction response from hash: %s transaction might be included in the next few blocks or check your node's health.", c, txHash)
			return

		}
	}
	l.Error(":anxious_face_with_sweat: Cannot send reports with adjusted gas: %d", c, gasLimit)
	return
}

// GetExecutable fetches data source executable using the provided client.
func GetExecutable(c *Context, l *Logger, hash string) ([]byte, error) {
	resValue, err := c.fileCache.GetFile(hash)
	if err != nil {
		l.Debug(":magnifying_glass_tilted_left: Fetching data source hash: %s from bandchain querier", hash)
		res, err := c.client.ABCIQuery(
			context.Background(),
			fmt.Sprintf("custom/%s/%s/%s", types.StoreKey, types.QueryData, hash),
			nil,
		)
		if err != nil {
			l.Error(":exploding_head: Failed to get data source with error: %s", c, err.Error())
			return nil, sdkerrors.Wrap(err, "failed to get data source")
		}
		resValue = res.Response.GetValue()
		c.fileCache.AddFile(resValue)
	} else {
		l.Debug(":card_file_box: Found data source hash: %s in cache file", hash)
	}

	l.Debug(":balloon: Received data source hash: %s content: %q", hash, resValue[:32])
	return resValue, nil
}

func GetDataSource(c *Context, l *Logger, id types.DataSourceID) (types.DataSource, error) {
	res, err := c.client.ABCIQuery(
		context.Background(),
		fmt.Sprintf("/store/%s/key", types.StoreKey),
		types.DataSourceStoreKey(id),
	)
	if err != nil {
		l.Debug(":skull: Failed to get data source with error: %s", err.Error())
		return types.DataSource{}, sdkerrors.Wrap(err, "failed to get data source")
	}

	var dataSource types.DataSource
	cdc.MustUnmarshalBinaryBare(res.Response.Value, &dataSource)

	_, _ = c.dataSourceCache.LoadOrStore(id, dataSource.Filename) // just put hash
	return dataSource, nil
}

// GetRequest fetches request by id
func GetRequest(c *Context, l *Logger, id types.RequestID) (types.Request, error) {
	res, err := c.client.ABCIQuery(
		context.Background(),
		fmt.Sprintf("/store/%s/key", types.StoreKey),
		types.RequestStoreKey(id),
	)
	if err != nil {
		l.Debug(":skull: Failed to get request with error: %s", err.Error())
		return types.Request{}, sdkerrors.Wrap(err, "failed to get request")
	}

	var r types.Request
	cdc.MustUnmarshalBinaryBare(res.Response.Value, &r)

	return r, nil
}
