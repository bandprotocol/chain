package yoda

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/GeoDB-Limited/odin-core/pkg/filecache"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/GeoDB-Limited/odin-core/yoda/executor"
)

type FeeEstimationData struct {
	askCount    int64
	minCount    int64
	callData    []byte
	rawRequests []rawRequest
	reports     []types.RawReport
	clientID    string
}

type ReportMsgWithKey struct {
	msg               *types.MsgReportData
	execVersion       []string
	keyIndex          int64
	feeEstimationData FeeEstimationData
}

type Context struct {
	client           rpcclient.Client
	validator        sdk.ValAddress
	gasPrices        string
	keys             []keyring.Info
	executor         executor.Executor
	fileCache        filecache.Cache
	broadcastTimeout time.Duration
	maxTry           uint64
	rpcPollInterval  time.Duration
	maxReport        uint64

	pendingMsgs        chan ReportMsgWithKey
	freeKeys           chan int64
	keyRoundRobinIndex int64 // Must use in conjunction with sync/atomic

	dataSourceCache *sync.Map
	pendingRequests map[types.RequestID]bool

	metricsEnabled bool
	handlingGauge  int64
	pendingGauge   int64
	errorCount     int64
	submittedCount int64
}

func (ctx *Context) nextKeyIndex() int64 {
	keyIndex := atomic.AddInt64(&ctx.keyRoundRobinIndex, 1) % int64(len(ctx.keys))
	return keyIndex
}

func (ctx *Context) updateHandlingGauge(amount int64) {
	if ctx.metricsEnabled {
		atomic.AddInt64(&ctx.handlingGauge, amount)
	}
}

func (ctx *Context) updatePendingGauge(amount int64) {
	if ctx.metricsEnabled {
		atomic.AddInt64(&ctx.pendingGauge, amount)
	}
}

func (ctx *Context) updateErrorCount(amount int64) {
	if ctx.metricsEnabled {
		atomic.AddInt64(&ctx.errorCount, amount)
	}
}

func (ctx *Context) updateSubmittedCount(amount int64) {
	if ctx.metricsEnabled {
		atomic.AddInt64(&ctx.submittedCount, amount)
	}
}
