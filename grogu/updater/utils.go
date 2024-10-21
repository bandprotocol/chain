package updater

import (
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/bandprotocol/chain/v3/pkg/logger"
)

func processEvent(
	ev coretypes.ResultEvent,
	logger *logger.Logger,
	query string,
	maxHeight *atomic.Int64,
	getHeight func(ev coretypes.ResultEvent) int64,
	updateFunc func() error,
) {
	height := getHeight(ev)
	if height > maxHeight.Load() {
		maxHeight.Store(height)
		logger.Info("[Updater] Received event for %s with new max height: %d", query, height)

		// Retry logic
		for {
			// If a new event has arrived, break the retry loop
			if maxHeight.Load() > height {
				logger.Debug("[Updater] A new event for %s arrived, aborting retry for height: %d", query, height)
				break
			}

			err := updateFunc()
			if err == nil {
				logger.Info("[Updater] Successfully processed event for %s at height: %d", query, height)
				break
			}

			// Check if the error is a gRPC error with code Status::not_found
			st, ok := status.FromError(err)
			if !ok || (st.Code() != codes.NotFound && st.Code() != codes.DeadlineExceeded) {
				logger.Error(
					"[Updater] Failed to process event for %s at height: %d with error: %v, not retrying",
					query,
					height,
					err,
				)
				break
			}

			logger.Debug(
				"[Updater] Failed to process event for %s at height: %d due to not found error, retrying...",
				query,
				height,
			)

			// Sleep for a while before retrying, to avoid hammering the system
			time.Sleep(1 * time.Second)
		}
	}
}
