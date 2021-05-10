package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

type LimitStatus struct {
	LastWithdrawal    time.Time
	WithdrawnInPeriod sdk.Coins
}

type Limit struct {
	cfg    Config
	ctx    *Context
	status *SyncMap
}

func NewLimit(ctx *Context, cfg Config) *Limit {
	return &Limit{
		cfg:    cfg,
		ctx:    ctx,
		status: NewSyncMap(),
	}
}

func (l *Limit) Allowed(rawAddress, denom string) (*LimitStatus, bool) {
	limitStatus, ok := l.status.Load(rawAddress)
	if !ok {
		return nil, true
	}

	if time.Now().Sub(limitStatus.LastWithdrawal) > l.cfg.Period {
		return limitStatus, true
	}

	if limitStatus.WithdrawnInPeriod.AmountOf(denom).LT(l.ctx.maxPerPeriodWithdrawal.AmountOf(denom)) {
		return limitStatus, true
	}
	return limitStatus, false
}
