package main

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	GasAmount     = 200000
	GasAdjustment = 1
)

// Request defines request of faucet withdrawal.
type Request struct {
	Denom   string `json:"denom" binding:"required"`
	Address string `json:"address" binding:"required"`
}

// Response defines response of faucet withdrawal.
type Response struct {
	TxHash string `json:"txHash"`
}

// HandleRequest handles faucet withdrawal.
func (l *Limiter) HandleRequest(gc *gin.Context) {
	key := <-l.keys
	defer func() {
		l.keys <- key
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
	if limitStatus, ok := l.allowed(req.Address, req.Denom); !ok {
		gc.JSON(
			http.StatusForbidden,
			gin.H{
				"error": "cannot withdraw more coins",
				"time":  (faucet.config.Period - time.Now().Sub(limitStatus.LastWithdrawals[req.Denom])).Seconds(),
			},
		)
		return
	}

	coinsToWithdraw := sdk.NewCoins(sdk.NewCoin(req.Denom, l.ctx.coins.AmountOf(req.Denom)))
	res, err := l.transferCoinsToClaimer(key, to, coinsToWithdraw)
	if err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if res.Code != 0 {
		gc.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": fmt.Sprintf(
					":exploding_head: Tx returned nonzero code %d with log %s, tx hash: %s",
					res.Code,
					res.RawLog,
					res.TxHash,
				),
			},
		)
		return
	}

	l.updateLimitation(req.Address, req.Denom, coinsToWithdraw)

	gc.JSON(http.StatusOK, Response{TxHash: res.TxHash})
}
