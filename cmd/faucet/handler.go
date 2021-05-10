package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gin-gonic/gin"

	band "github.com/GeoDB-Limited/odin-core/app"
)

type Request struct {
	Denom   string `json:"denom" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type Response struct {
	TxHash string `json:"txHash"`
}

var (
	cdc, _ = band.MakeCodecs()
)

func handleRequest(gc *gin.Context, c *Context) {
	key := <-c.keys
	defer func() {
		c.keys <- key
	}()

	var req Request
	if err := gc.ShouldBindJSON(&req); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	to, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sdk.ValidateDenom(req.Denom); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if limitStatus, ok := limit.Allowed(req.Address, req.Denom); !ok {
		gc.JSON(http.StatusBadRequest, gin.H{"error": "cannot withdraw more coins", "time": (cfg.Period - time.Now().Sub(limitStatus.LastWithdrawal)).String()})
		return
	}
	coinsToWithdraw := sdk.NewCoins(sdk.NewCoin(req.Denom, c.coins.AmountOf(req.Denom)))
	msg := banktypes.NewMsgSend(key.GetAddress(), to, coinsToWithdraw)
	if err := msg.ValidateBasic(); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientCtx := client.Context{
		Client:            c.client,
		TxConfig:          band.MakeEncodingConfig().TxConfig,
		BroadcastMode:     "async",
		InterfaceRegistry: band.MakeEncodingConfig().InterfaceRegistry,
	}
	accountRetriever := authtypes.AccountRetriever{}
	acc, err := accountRetriever.GetAccount(clientCtx, key.GetAddress())
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(band.MakeEncodingConfig().TxConfig).
		WithGas(200000).WithGasAdjustment(1).
		WithChainID(cfg.ChainID).
		WithMemo("").
		WithGasPrices(c.gasPrices.String()).
		WithKeybase(keybase).
		WithAccountRetriever(clientCtx.AccountRetriever)

	txb, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = tx.Sign(txf, key.GetName(), txb, true)
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
	res, err := clientCtx.BroadcastTxCommit(txBytes)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.Code != 0 {
		gc.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf(":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s",
				res.Code, res.RawLog, res.TxHash,
			)})
		return
	}

	limitStatus, ok := limit.status.Load(req.Address)
	if !ok {
		limitStatus = &LimitStatus{
			LastWithdrawal:    time.Now(),
			WithdrawnInPeriod: sdk.NewCoins(),
		}
	}
	limitStatus.WithdrawnInPeriod = limitStatus.WithdrawnInPeriod.Add(coinsToWithdraw...)
	limit.status.Store(req.Address, limitStatus)
	gc.JSON(200, Response{
		TxHash: res.TxHash,
	})

}
