package limiter

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/cmd/faucet/store"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	odin "github.com/GeoDB-Limited/odin-core/app"
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
	if limitStatus, ok := l.Allowed(req.Address, req.Denom); !ok {
		gc.JSON(
			http.StatusForbidden,
			gin.H{
				"error": "cannot withdraw more coins",
				"time":  (l.cfg.Period - time.Now().Sub(limitStatus.LastWithdrawals[req.Denom])).Seconds(),
			},
		)
		return
	}
	coinsToWithdraw := sdk.NewCoins(sdk.NewCoin(req.Denom, l.cfg.Coins.AmountOf(req.Denom)))
	msg := banktypes.NewMsgSend(key.GetAddress(), to, coinsToWithdraw)
	if err := msg.ValidateBasic(); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientCtx := client.Context{
		Client:            l.client,
		TxConfig:          odin.MakeEncodingConfig().TxConfig,
		BroadcastMode:     flags.BroadcastAsync,
		InterfaceRegistry: odin.MakeEncodingConfig().InterfaceRegistry,
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
		WithTxConfig(odin.MakeEncodingConfig().TxConfig).
		WithGas(GasAmount).WithGasAdjustment(GasAdjustment).
		WithChainID(l.cfg.ChainID).
		WithMemo("").
		WithGasPrices(l.cfg.GasPrices.String()).
		WithKeybase(l.cfg.Keyring).
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

	withdrawalLimit, ok := l.store.Get(req.Address)
	if !ok {
		withdrawalLimit = &store.WithdrawalLimit{
			LastWithdrawals:  make(map[string]time.Time),
			WithdrawalPeriod: sdk.NewCoins(),
		}
	}
	withdrawalLimit.LastWithdrawals[req.Denom] = time.Now()
	withdrawalLimit.WithdrawalPeriod = withdrawalLimit.WithdrawalPeriod.Add(coinsToWithdraw...)
	l.store.Set(req.Address, withdrawalLimit)

	gc.JSON(http.StatusOK, Response{TxHash: res.TxHash})
}
