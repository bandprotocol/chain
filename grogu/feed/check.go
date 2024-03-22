package feed

import (
	"context"
	"time"

	"golang.org/x/exp/maps"

	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func checkFeeds(c *grogucontext.Context, l *grogucontext.Logger) {
	validValidator, err := c.QueryClient.ValidValidator(context.Background(), &types.QueryValidValidatorRequest{
		Validator: c.Validator.String(),
	})
	if err != nil {
		return
	}

	if !validValidator.Valid {
		return
	}

	paramsResponse, err := c.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return
	}
	params := paramsResponse.Params

	feedsResponse, err := c.QueryClient.SupportedFeeds(context.Background(), &types.QuerySupportedFeedsRequest{})
	if err != nil {
		return
	}

	feeds := feedsResponse.Feeds

	validatorPricesResponse, err := c.QueryClient.ValidatorPrices(
		context.Background(),
		&types.QueryValidatorPricesRequest{
			Validator: c.Validator.String(),
		},
	)
	if err != nil {
		return
	}

	validatorPrices := validatorPricesResponse.ValidatorPrices
	signalIDTimestampMap := convertToSignalIDTimestampMap(validatorPrices)

	requestedSignalIDs := make(map[string]time.Time)
	now := time.Now()

	for _, feed := range feeds {
		if _, inProgress := c.InProgressSignalIDs.Load(feed.SignalID); inProgress {
			continue
		}

		timestamp, ok := signalIDTimestampMap[feed.SignalID]
		// add 2 to prevent too fast cases
		if !ok ||
			time.Unix(timestamp+2, 0).
				Add(time.Duration(feed.Interval)*time.Second).
				Add(-time.Duration(params.TransitionTime)*time.Second).
				Before(now) {
			requestedSignalIDs[feed.SignalID] = time.Unix(timestamp, 0).
				Add(time.Duration(feed.Interval) * time.Second).
				Add(-time.Duration(params.TransitionTime) * time.Second / 2)
			c.InProgressSignalIDs.Store(feed.SignalID, time.Now())
		}
	}
	if len(requestedSignalIDs) != 0 {
		l.Info("found signal ids to send: %v", maps.Keys(requestedSignalIDs))
		c.PendingSignalIDs <- requestedSignalIDs
	}
}

// convertToSignalIDTimestampMap converts an array of PriceValidator to a map of signal id to timestamp.
func convertToSignalIDTimestampMap(data []types.PriceValidator) map[string]int64 {
	signalIDTimestampMap := make(map[string]int64)

	for _, entry := range data {
		signalIDTimestampMap[entry.SignalID] = entry.Timestamp
	}

	return signalIDTimestampMap
}

func StartCheckFeeds(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		checkFeeds(c, l)
		time.Sleep(time.Second)
	}
}
