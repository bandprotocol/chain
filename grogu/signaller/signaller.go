package signaller

import (
	"math"
	"sync"
	"time"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client"
	proto "github.com/bandprotocol/bothan/bothan-api/client/go-client/query"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/grogu/querier"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	Multiplier                = 1_000_000_000
	UpperBound                = float64(math.MaxUint64) / Multiplier
	FixedIntervalOffset int64 = 10
	TimeBuffer          int64 = 3
)

type Signaller struct {
	feedQuerier  *querier.FeedQuerier
	bothanClient bothan.Client
	// How often to check for signal changes
	interval         time.Duration
	submitCh         chan<- []types.SignalPrice
	logger           *logger.Logger
	valAddress       sdk.ValAddress
	pendingSignalIDs *sync.Map

	distributionStartPercentage  uint64
	distributionOffsetPercentage uint64

	signalIDToFeed           map[string]types.FeedWithDeviation
	signalIDToValidatorPrice map[string]types.ValidatorPrice
	params                   *types.Params
}

func New(
	feedQuerier *querier.FeedQuerier,
	bothanClient bothan.Client,
	interval time.Duration,
	submitCh chan<- []types.SignalPrice,
	logger *logger.Logger,
	valAddress sdk.ValAddress,
	pendingSignalIDs *sync.Map,
	distributionStartPercentage uint64,
	distributionOffsetPercentage uint64,
) *Signaller {
	return &Signaller{
		feedQuerier:                  feedQuerier,
		bothanClient:                 bothanClient,
		interval:                     interval,
		submitCh:                     submitCh,
		logger:                       logger,
		valAddress:                   valAddress,
		pendingSignalIDs:             pendingSignalIDs,
		distributionStartPercentage:  distributionStartPercentage,
		distributionOffsetPercentage: distributionOffsetPercentage,
		signalIDToFeed:               make(map[string]types.FeedWithDeviation),
		signalIDToValidatorPrice:     make(map[string]types.ValidatorPrice),
		params:                       nil,
	}
}

func (h *Signaller) Start() {
	for {
		time.Sleep(h.interval)

		resp, err := h.feedQuerier.QueryValidValidator(h.valAddress)
		if err != nil {
			h.logger.Error("[Signaller] failed to query valid validator: %v", err)
			continue
		}

		if !resp.Valid {
			h.logger.Info("[Signaller] validator is not required to feed prices")
			continue
		}

		if !h.updateInternalVariables() {
			h.logger.Error("[Signaller] failed to update internal variables: %v", err)
			continue
		}

		h.execute()
	}
}

func (h *Signaller) updateInternalVariables() bool {
	resultCh := make(chan bool, 3)
	var wg sync.WaitGroup

	updater := func(f func() bool) {
		defer wg.Done()
		resultCh <- f()
	}

	wg.Add(3)
	go updater(h.updateParams)
	go updater(h.updateFeedMap)
	go updater(h.updateValidatorPriceMap)
	wg.Wait()
	close(resultCh)

	success := true
	for result := range resultCh {
		if !result {
			success = false
			break
		}
	}

	return success
}

func (h *Signaller) updateParams() bool {
	resp, err := h.feedQuerier.QueryParams()
	if err != nil {
		h.logger.Error("[Signaller] failed to query params: %v", err)
		return false
	}

	h.params = &resp.Params
	return true
}

func (h *Signaller) updateFeedMap() bool {
	resp, err := h.feedQuerier.QuerySupportedFeeds()
	if err != nil {
		h.logger.Error("[Signaller] failed to query supported feeds: %v", err)
		return false
	}

	h.signalIDToFeed = sliceToMap(resp.SupportedFeeds.Feeds, func(feed types.FeedWithDeviation) string {
		return feed.SignalID
	})

	return true
}

func (h *Signaller) updateValidatorPriceMap() bool {
	resp, err := h.feedQuerier.QueryValidatorPrices(h.valAddress)
	if err != nil {
		h.logger.Error("[Signaller] failed to query validator prices: %v", err)
		return false
	}

	h.signalIDToValidatorPrice = sliceToMap(resp.ValidatorPrices, func(valPrice types.ValidatorPrice) string {
		return valPrice.SignalID
	})

	return true
}

func (h *Signaller) execute() {
	now := time.Now().Unix()

	h.logger.Debug("[Signaller] starting signal process")
	signalIDs := make([]string, 0, len(h.signalIDToFeed))
	for signalID := range h.signalIDToFeed {
		signalIDs = append(signalIDs, signalID)
	}

	h.logger.Debug("[Signaller] filtering signal ids")
	nonPendingSignalIDs := h.filterPendingSignalIDs(signalIDs)
	if len(nonPendingSignalIDs) == 0 {
		h.logger.Debug("[Signaller] no signal ids to process")
		return
	}

	h.setPendingSignalIDs(nonPendingSignalIDs)

	h.logger.Debug("[Signaller] querying prices from bothan")
	prices, err := h.bothanClient.QueryPrices(nonPendingSignalIDs)
	if err != nil {
		h.logger.Error("[Signaller] failed to query prices from bothan: %v", err)
		h.removePendingSignalIDs(nonPendingSignalIDs)
		return
	}

	h.logger.Debug("[Signaller] filtering prices")
	pricesMap := sliceToMap(prices, func(price *proto.PriceData) string {
		return price.SignalId
	})
	filteredPrices := h.filterPrices(pricesMap, nonPendingSignalIDs)
	submitPrices, unusedSignalIDs := h.prepareSubmitPrices(filteredPrices, now)

	h.removePendingSignalIDs(unusedSignalIDs)

	if len(submitPrices) == 0 {
		h.logger.Debug("[Signaller] no prices to submit")
		return
	}
	h.logger.Debug("[Signaller] submitting prices: %v", submitPrices)

	h.submitCh <- submitPrices
}

func (h *Signaller) setPendingSignalIDs(signalIDs []string) {
	for _, signalID := range signalIDs {
		h.pendingSignalIDs.Store(signalID, struct{}{})
	}
}

func (h *Signaller) removePendingSignalIDs(signalIDs []string) {
	for _, signalID := range signalIDs {
		h.pendingSignalIDs.Delete(signalID)
	}
}

func (h *Signaller) filterPendingSignalIDs(signalIDs []string) []string {
	filtered := make([]string, 0, len(signalIDs))
	for _, signalID := range signalIDs {
		if _, ok := h.pendingSignalIDs.Load(signalID); !ok {
			filtered = append(filtered, signalID)
		}
	}
	return filtered
}

func (h *Signaller) filterPrices(
	pricesMap map[string]*proto.PriceData,
	signalIDs []string,
) []*proto.PriceData {
	now := time.Now()
	toUpdatePrices := make([]*proto.PriceData, 0, len(signalIDs))

	for _, signalID := range signalIDs {
		price, ok := pricesMap[signalID]
		if !ok {
			// If the price is not found, remove it from the pending signal IDs
			h.logger.Debug("[Signaller] price not found for signal ID: %s", signalID)
			h.pendingSignalIDs.Delete(signalID)
			continue
		}

		if h.isPriceValid(price, now) {
			toUpdatePrices = append(toUpdatePrices, price)
		} else {
			// If the price is not valid, remove it from the pending signal IDs
			h.pendingSignalIDs.Delete(price.SignalId)
		}
	}

	return toUpdatePrices
}

func (h *Signaller) isPriceValid(
	price *proto.PriceData,
	now time.Time,
) bool {
	// Check if the price is supported and required to be submitted
	feed, ok := h.signalIDToFeed[price.SignalId]
	if !ok {
		return false
	}

	// If unable to convert price, it is considered invalid
	newPrice, err := safeConvert(price.Price)
	if err != nil {
		h.logger.Error("[Signaller] failed to convert price: %v", err)
		return false
	}

	// Get the last price submitted by the validator, if it doesn't exist, it is valid to be sent
	valPrice, ok := h.signalIDToValidatorPrice[price.SignalId]
	if !ok {
		return true
	}

	// If the last price exists, check if the price can be updated
	if h.shouldUpdatePrice(feed, valPrice, newPrice, now) {
		return true
	}

	return false
}

func (h *Signaller) shouldUpdatePrice(
	feed types.FeedWithDeviation,
	valPrice types.ValidatorPrice,
	newPrice uint64,
	now time.Time,
) bool {
	// thresholdTime is the time when the price can be updated.
	// add TimeBuffer to make sure the thresholdTime is not too early.
	thresholdTime := time.Unix(valPrice.Timestamp+h.params.CooldownTime+TimeBuffer, 0)

	if thresholdTime.After(now) {
		return false
	}

	// Check if the price is past the assigned time, if it is, add it to the list of prices to update
	assignedTime := calculateAssignedTime(
		h.valAddress,
		feed.Interval,
		valPrice.Timestamp,
		h.distributionOffsetPercentage,
		h.distributionStartPercentage,
	)

	if assignedTime.Before(now) {
		return true
	}

	// Check if the price is deviated from the last submission, if it is, add it to the list of prices to update
	if isDeviated(feed.DeviationBasisPoint, valPrice.Price, newPrice) {
		return true
	}

	return false
}

func (h *Signaller) prepareSubmitPrices(filteredPrices []*proto.PriceData, now int64) ([]types.SignalPrice, []string) {
	submitPrices := make([]types.SignalPrice, 0, len(filteredPrices))
	unusedSignalIDs := make([]string, 0, len(filteredPrices))

	for _, price := range filteredPrices {
		submitPrice, err := convertPriceData(price)
		if err != nil {
			h.logger.Debug("[Signaller] failed to parse price data: %v", err)
			unusedSignalIDs = append(unusedSignalIDs, price.SignalId)
			continue
		}

		switch submitPrice.PriceStatus {
		case types.PriceStatusUnavailable:
			deadline := h.signalIDToValidatorPrice[price.SignalId].Timestamp + h.signalIDToFeed[price.SignalId].Interval
			if now > deadline-FixedIntervalOffset {
				submitPrices = append(submitPrices, submitPrice)
			} else {
				unusedSignalIDs = append(unusedSignalIDs, price.SignalId)
			}
		default:
			submitPrices = append(submitPrices, submitPrice)
		}
	}

	return submitPrices, unusedSignalIDs
}
