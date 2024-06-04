package feed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	bothanproto "github.com/bandprotocol/bothan/bothan-api/client/go-client/query"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/maps"

	grogucontext "github.com/bandprotocol/chain/v2/grogu/context"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func checkFeeds(c *grogucontext.Context) {
	// Fetch parameters, supported feeds, validator prices, and prices
	params, feeds, validatorPrices, _, err := fetchData(c)
	if err != nil {
		return
	}

	signalIDTimestampMap := convertToSignalIDTimestampMap(validatorPrices)
	signalIDValidatorPriceMap := convertToSignalIDValidatorPriceMap(validatorPrices)

	requestedSignalIDs := make(map[string]time.Time)
	now := time.Now()

	for _, feed := range feeds {
		// Skip feeds in progress
		if _, inProgress := c.InProgressSignalIDs.Load(feed.SignalID); inProgress {
			continue
		}

		// Get latest timestamp of the feed
		timestamp, ok := signalIDTimestampMap[feed.SignalID]
		// If there is no timestamp yet, then puts it in requested signal id list
		if !ok {
			updateRequestedSignalID(c, requestedSignalIDs, feed, timestamp, params)
			continue
		}

		// Skip if it's in cooldown
		if time.Unix(timestamp+2, 0).
			Add(time.Duration(params.CooldownTime) * time.Second).
			After(now) {
			continue
		}

		// Calculate assigned time for the feed
		assignedTime := calculateAssignedTime(
			c.Validator,
			feed.Interval,
			timestamp,
			c.Config.DistributionPercentageRange,
			c.Config.DistributionStartPercentage,
		)

		if assignedTime.Before(now) || isDeviate(c, feed, signalIDValidatorPriceMap) {
			updateRequestedSignalID(c, requestedSignalIDs, feed, timestamp, params)
		}
	}

	if len(requestedSignalIDs) != 0 {
		c.Logger.Info("found signal ids to send: %v", maps.Keys(requestedSignalIDs))
		c.PendingSignalIDs <- requestedSignalIDs
	}
}

func fetchData(
	c *grogucontext.Context,
) (params types.Params, feeds []types.Feed, validatorPrices []types.ValidatorPrice, prices []*types.Price, err error) {
	// Fetch validator data
	validValidator, err := c.QueryClient.ValidValidator(context.Background(), &types.QueryValidValidatorRequest{
		Validator: c.Validator.String(),
	})
	if err != nil {
		return types.Params{}, nil, nil, nil, err
	}
	if !validValidator.Valid {
		return types.Params{}, nil, nil, nil, fmt.Errorf("validator is not valid or not required to send price")
	}

	// Fetch params
	paramsResponse, err := c.QueryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return types.Params{}, nil, nil, nil, err
	}
	params = paramsResponse.Params

	// Fetch supported feeds
	feedsResponse, err := c.QueryClient.SupportedFeeds(context.Background(), &types.QuerySupportedFeedsRequest{})
	if err != nil {
		return types.Params{}, nil, nil, nil, err
	}
	feeds = feedsResponse.SupportedFeeds.Feeds

	// Fetch validator prices
	validatorPricesResponse, err := c.QueryClient.ValidatorPrices(
		context.Background(),
		&types.QueryValidatorPricesRequest{
			Validator: c.Validator.String(),
		},
	)
	if err != nil {
		return types.Params{}, nil, nil, nil, err
	}
	validatorPrices = validatorPricesResponse.ValidatorPrices

	// Fetch prices
	pricesResponse, err := c.QueryClient.Prices(
		context.Background(),
		&types.QueryPricesRequest{},
	)
	if err != nil {
		return types.Params{}, nil, nil, nil, err
	}
	prices = pricesResponse.Prices

	return params, feeds, validatorPrices, prices, nil
}

// calculateAssignedTime calculates the assigned time for the feed
func calculateAssignedTime(
	valAddr sdk.ValAddress,
	interval int64,
	timestamp int64,
	dpRange uint64,
	dpStart uint64,
) time.Time {
	hashed := sha256.Sum256(append(valAddr.Bytes(), sdk.Uint64ToBigEndian(uint64(timestamp))...))
	offset := sdk.BigEndianToUint64(
		hashed[:],
	)%dpRange + dpStart
	timeOffset := interval * int64(offset) / 100
	// add 2 seconds to prevent too fast case
	return time.Unix(timestamp+2, 0).Add(time.Duration(timeOffset) * time.Second)
}

// isDeviate checks if the current price is deviated from the on-chain validator price
func isDeviate(
	c *grogucontext.Context,
	feed types.Feed,
	signalIDValidatorPriceMap map[string]uint64,
) bool {
	currentPrices, err := c.PriceService.Query([]string{feed.SignalID})
	if err != nil || len(currentPrices) == 0 ||
		currentPrices[0].PriceStatus != bothanproto.PriceStatus_PRICE_STATUS_AVAILABLE {
		return false
	}

	price, err := strconv.ParseFloat(strings.TrimSpace(currentPrices[0].Price), 64)
	if err != nil {
		return false
	}

	if signalIDValidatorPriceMap[feed.SignalID] == 0 {
		return true
	}

	return feed.DeviationInThousandth <= deviationInThousandth(
		signalIDValidatorPriceMap[feed.SignalID],
		uint64(price*math.Pow10(9)),
	)
}

// updateRequestedSignalID updates the requestedSignalIDs map and stores the timestamp in InProgressSignalIDs for a feed
func updateRequestedSignalID(
	c *grogucontext.Context,
	requestedSignalIDs map[string]time.Time,
	feed types.Feed,
	timestamp int64,
	params types.Params,
) {
	requestedSignalIDs[feed.SignalID] = time.Unix(timestamp, 0).
		Add(time.Duration(feed.Interval) * time.Second).
		Add(-time.Duration(params.TransitionTime) * time.Second / 2)
	c.InProgressSignalIDs.Store(feed.SignalID, time.Now())
}

// convertToSignalIDTimestampMap converts an array of ValidatorPrice to a map of signal id to timestamp.
func convertToSignalIDTimestampMap(data []types.ValidatorPrice) map[string]int64 {
	signalIDTimestampMap := make(map[string]int64)

	for _, entry := range data {
		signalIDTimestampMap[entry.SignalID] = entry.Timestamp
	}

	return signalIDTimestampMap
}

// convertToSignalIDValidatorPriceMap converts an array of Prices to a map of signal id to its on-chain validator prices.
func convertToSignalIDValidatorPriceMap(data []types.ValidatorPrice) map[string]uint64 {
	signalIDValidatorPriceMap := make(map[string]uint64)

	for _, entry := range data {
		signalIDValidatorPriceMap[entry.SignalID] = entry.Price
	}

	return signalIDValidatorPriceMap
}

// deviationInThousandth calculates the deviation in thousandth between two values.
func deviationInThousandth(originalValue, newValue uint64) int64 {
	diff := math.Abs(float64(newValue) - float64(originalValue))
	deviation := (diff / float64(originalValue)) * 1000
	return int64(deviation)
}

func StartCheckFeeds(c *grogucontext.Context) {
	for {
		checkFeeds(c)
		time.Sleep(time.Second)
	}
}
