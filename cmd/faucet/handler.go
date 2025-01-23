package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Request struct {
	Address string `json:"address" binding:"required"`
}

type Response struct {
	TxHash string `json:"txHash"`
}

func handleRequest(gc *gin.Context, c *Context) {
	key := <-c.keys
	defer func() {
		c.keys <- key
	}()

	var req Request
	if err := gc.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	to, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address, err := key.GetAddress()
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := banktypes.NewMsgSend(address, to, c.amount)

	clientCtx := client.Context{
		Client:            c.client,
		Codec:             c.encodingConfig.Codec,
		TxConfig:          c.encodingConfig.TxConfig,
		BroadcastMode:     "async",
		InterfaceRegistry: c.encodingConfig.InterfaceRegistry,
	}
	accountRetriever := authtypes.AccountRetriever{}
	acc, err := accountRetriever.GetAccount(clientCtx, address)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(c.encodingConfig.TxConfig).
		WithGas(200000).WithGasAdjustment(1).
		WithChainID(cfg.ChainID).
		WithMemo("").
		WithGasPrices(c.gasPrices.String()).
		WithKeybase(keybase).
		WithAccountRetriever(clientCtx.AccountRetriever)

	txb, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = tx.Sign(context.Background(), txf, key.Name, txb, true)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTxSync(txBytes)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.Code != 0 {
		gc.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf(":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s",
				res.Code, res.RawLog, res.TxHash,
			),
		})
		return
	}
	gc.JSON(200, Response{
		TxHash: res.TxHash,
	})
}
