package utils

import (
	"github.com/bandprotocol/chain/v3/cylinder/msg"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

const MAX_RETRY = 3

// CheckResultAndRetry checks the result of the response and retries the request if it fails.
func CheckResultAndRetry(
	logger *logger.Logger,
	resp msg.Response,
	msgCh chan msg.Request,
	msgName string,
) {
	req := resp.Request
	if resp.Success {
		logger.Info(
			":smiling_face_with_sunglasses: Successfully sent %s ID: %d, txHash: %s",
			msgName,
			req.ID,
			resp.TxHash,
		)
	} else {
		if req.Retry < MAX_RETRY {
			logger.Debug(
				":anxious_face_with_sweat: Failed to send %s ID: %d, retry: %d; %s",
				msgName,
				req.ID,
				req.Retry,
				resp.Err,
			)

			logger.Info(
				":delivery_truck: Retry request ID: %d, retry: %d",
				req.ID,
				req.Retry+1,
			)

			msgCh <- req.IncreaseRetry()
		} else {
			logger.Error(
				":anxious_face_with_sweat: Failed to send %s request ID: %d, retry: %d; %s",
				msgName,
				req.ID,
				req.Retry,
				resp.Err,
			)
		}
	}
}
