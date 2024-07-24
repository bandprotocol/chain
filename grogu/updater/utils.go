package updater

import (
	"sync/atomic"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/bandprotocol/chain/v2/pkg/logger"
)

func processEvent(
	ev coretypes.ResultEvent,
	logger *logger.Logger,
	query string,
	maxHeight *atomic.Int64,
	getHeight func(ev coretypes.ResultEvent) int64,
	updateFunc func(),
) {
	height := getHeight(ev)
	if height > maxHeight.Load() {
		maxHeight.Store(height)
		logger.Debug("[Updater] Received event for %s with new max height: %d", query, height)
		go updateFunc()
	}
}
