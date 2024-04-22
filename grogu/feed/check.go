package feed

import (
	"context"
	"crypto/sha256"
	"math"
	"strconv"
	"strings"
	"time"

	bothanproto "github.com/bandprotocol/bothan-api/go-proxy/proto"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	pricesResponse, err := c.QueryClient.Prices(
		context.Background(),
		&types.QueryPricesRequest{},
	)
	if err != nil {
		return
	}

	signalIDChainPriceMap := convertToSignalIDChainPriceMap(pricesResponse.Prices)

	requestedSignalIDs := make(map[string]time.Time)
	now := time.Now()

	for _, feed := range feeds {
		if _, inProgress := c.InProgressSignalIDs.Load(feed.SignalID); inProgress {
			continue
		}

		timestamp, ok := signalIDTimestampMap[feed.SignalID]
		if !ok {
			requestedSignalIDs[feed.SignalID] = time.Unix(timestamp, 0).
				Add(time.Duration(feed.Interval) * time.Second).
				Add(-time.Duration(params.TransitionTime) * time.Second / 2)
			c.InProgressSignalIDs.Store(feed.SignalID, time.Now())
			continue
		}

		// hash validator address and timestamp of last price submission
		hashed := sha256.Sum256(append([]byte(c.Validator), sdk.Uint64ToBigEndian(uint64(timestamp))...))

		// calculate a time offset for next price submission
		offset := sdk.BigEndianToUint64(hashed[:])%30 + 50
		time_offset := feed.Interval * int64(offset) / 100

		// calculate next assigned time for this signal id
		// add 2 to prevent too fast cases
		assigned_time := time.Unix(timestamp+2, 0).
			Add(time.Duration(time_offset) * time.Second)

		if assigned_time.Before(now) {
			requestedSignalIDs[feed.SignalID] = time.Unix(timestamp, 0).
				Add(time.Duration(feed.Interval) * time.Second).
				Add(-time.Duration(params.TransitionTime) * time.Second / 2)
			c.InProgressSignalIDs.Store(feed.SignalID, time.Now())
			continue
		}

		if time.Unix(timestamp+2, 0).
			Add(time.Duration(params.CooldownTime) * time.Second).
			Before(now) {
			currentPrices, err := c.PriceService.Query([]string{feed.SignalID})
			if err != nil || len(currentPrices) == 0 ||
				currentPrices[0].PriceOption != bothanproto.PriceOption_PRICE_OPTION_AVAILABLE {
				continue
			}

			price, err := strconv.ParseFloat(strings.TrimSpace(currentPrices[0].Price), 64)
			if err != nil {
				continue
			}

			if feed.DeviationInThousandth <= deviationInThousandth(
				signalIDChainPriceMap[feed.SignalID],
				uint64(price*math.Pow10(9)),
			) {
				requestedSignalIDs[feed.SignalID] = time.Unix(timestamp, 0).
					Add(time.Duration(feed.Interval) * time.Second).
					Add(-time.Duration(params.TransitionTime) * time.Second / 2)
				c.InProgressSignalIDs.Store(feed.SignalID, time.Now())
				continue
			}
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

// convertToSignalIDChainPriceMap converts an array of Prices to a map of signal id to its on-chain prices.
func convertToSignalIDChainPriceMap(data []*types.Price) map[string]uint64 {
	signalIDChainPriceMap := make(map[string]uint64)

	for _, entry := range data {
		signalIDChainPriceMap[entry.SignalID] = entry.Price
	}

	return signalIDChainPriceMap
}

// deviationInThousandth calculates the deviation in thousandth between two values.
func deviationInThousandth(originalValue, newValue uint64) int64 {
	diff := math.Abs(float64(newValue - originalValue))
	deviation := (diff / float64(originalValue)) * 1000
	return int64(deviation)
}

func StartCheckFeeds(c *grogucontext.Context, l *grogucontext.Logger) {
	for {
		checkFeeds(c, l)
		time.Sleep(time.Second)
	}
}
