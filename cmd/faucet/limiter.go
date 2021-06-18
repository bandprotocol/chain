package main

import (
	odin "github.com/GeoDB-Limited/odin-core/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

const (
	TickerUpdatePeriod = 30 * time.Second
)

// Limiter defines service for limiting faucet withdrawals.
type Limiter struct {
	ctx    *Context
	store  LimitStore
	client rpcclient.Client
	ticker *time.Ticker
	keys   chan keyring.Info
}

// NewLimiter creates a new limiter.
func NewLimiter(ctx *Context) *Limiter {
	httpClient, err := httpclient.New(faucet.config.NodeURI, "/websocket")
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create http client"))
	}

	kb, err := faucet.keybase.List()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to retrieve keys from keybase"))
	}
	if len(kb) == 0 {
		panic(sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "there are no available keys"))
	}
	keys := make(chan keyring.Info, len(kb))
	for _, key := range kb {
		keys <- key
	}

	return &Limiter{
		ctx:    ctx,
		store:  NewLimitStore(),
		client: httpClient,
		ticker: time.NewTicker(TickerUpdatePeriod),
		keys:   keys,
	}
}

// runCleaner removes deprecated limits per period.
func (l *Limiter) runCleaner() {
	for {
		select {
		case <-l.ticker.C:
			l.store.Clean(faucet.config.Period)
		}
	}
}

// allowed implements Limiter interface.
func (l *Limiter) allowed(rawAddress, denom string) (*WithdrawalLimit, bool) {
	limit, ok := l.store.Get(rawAddress)
	if !ok {
		return nil, true
	}
	if time.Now().Sub(limit.LastWithdrawals[denom]) > faucet.config.Period {
		return limit, true
	}
	if limit.WithdrawalAmount.AmountOf(denom).LT(l.ctx.maxPerPeriodWithdrawal.AmountOf(denom)) {
		return limit, true
	}
	return limit, false
}

// updateLimitation updates the limitations of account by the given address.
func (l *Limiter) updateLimitation(address, denom string, coins sdk.Coins) {
	withdrawalLimit, ok := l.store.Get(address)
	if !ok {
		withdrawalLimit = &WithdrawalLimit{
			LastWithdrawals:  make(map[string]time.Time),
			WithdrawalAmount: sdk.NewCoins(),
		}
	}
	withdrawalLimit.LastWithdrawals[denom] = time.Now()
	withdrawalLimit.WithdrawalAmount = withdrawalLimit.WithdrawalAmount.Add(coins...)
	l.store.Set(address, withdrawalLimit)
}

// transferCoinsToClaimer transfers coins from faucet accounts to the claimer.
func (l *Limiter) transferCoinsToClaimer(key keyring.Info, to sdk.AccAddress, amt sdk.Coins) (*sdk.TxResponse, error) {
	msg := banktypes.NewMsgSend(key.GetAddress(), to, amt)
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to validate a message: %s", msg.String())
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
		return nil, sdkerrors.Wrapf(err, "failed to the account: %s", acc)
	}

	txf := tx.Factory{}.
		WithAccountNumber(acc.GetAccountNumber()).
		WithSequence(acc.GetSequence()).
		WithTxConfig(odin.MakeEncodingConfig().TxConfig).
		WithGas(GasAmount).WithGasAdjustment(GasAdjustment).
		WithChainID(faucet.config.ChainID).
		WithMemo("").
		WithGasPrices(l.ctx.gasPrices.String()).
		WithKeybase(faucet.keybase).
		WithAccountRetriever(clientCtx.AccountRetriever)

	txb, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to build unsigned tx")
	}

	err = tx.Sign(txf, key.GetName(), txb, true)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to sign tx")
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to encode tx")
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTxCommit(txBytes)
	return res, sdkerrors.Wrap(err, "failed to broadcast tx commit")
}
