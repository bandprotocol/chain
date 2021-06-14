package limiter

import (
	"github.com/GeoDB-Limited/odin-core/cmd/faucet/config"
	store2 "github.com/GeoDB-Limited/odin-core/cmd/faucet/store"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

const (
	TickerUpdatePeriod = 30 * time.Second
)

// Limiter defines service for limiting faucet withdrawals.
type Limiter struct {
	cfg    *config.Config
	store  store2.LimitStore
	client rpcclient.Client
	ticker *time.Ticker
	keys   chan keyring.Info
}

// NewLimiter creates a new limiter.
func NewLimiter(cfg *config.Config) *Limiter {
	keyringList, err := cfg.Keyring.List()
	if err != nil {
		panic(err)
	}
	if len(keyringList) == 0 {
		panic("no key available")
	}
	keys := make(chan keyring.Info, len(keyringList))
	for _, key := range keyringList {
		keys <- key
	}

	client, err := httpclient.New(cfg.NodeURI, "/websocket")
	if err != nil {
		panic(err)
	}

	return &Limiter{
		cfg:    cfg,
		store:  store2.NewLimitStore(),
		client: client,
		ticker: time.NewTicker(TickerUpdatePeriod),
		keys:   keys,
	}
}

// RunCleaner removes deprecated limits per period.
func (l *Limiter) RunCleaner() {
	for {
		select {
		case <-l.ticker.C:
			l.store.Clean(l.cfg.Period)
		}
	}
}

// Allowed implements Limiter interface.
func (l *Limiter) Allowed(rawAddress, denom string) (*store2.WithdrawalLimit, bool) {
	limit, ok := l.store.Get(rawAddress)
	if !ok {
		return nil, true
	}
	if time.Now().Sub(limit.LastWithdrawals[denom]) > l.cfg.Period {
		return limit, true
	}
	if limit.WithdrawalPeriod.AmountOf(denom).LT(l.cfg.MaxWithdrawalPerPeriod.AmountOf(denom)) {
		return limit, true
	}
	return limit, false
}
