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
	resp, err := h.feedQuerier.QueryCurrentFeeds()
	if err != nil {
		h.logger.Error("[Signaller] failed to query supported feeds: %v", err)
		return false
	}

	h.signalIDToFeed = sliceToMap(resp.CurrentFeeds.Feeds, func(feed types.FeedWithDeviation) string {
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
	now := time.Now()

	h.logger.Debug("[Signaller] starting signal process")

	h.logger.Debug("[Signaller] getting non-pending signal ids")
	nonPendingSignalIDs := h.getNonPendingSignalIDs()
	if len(nonPendingSignalIDs) == 0 {
		h.logger.Debug("[Signaller] no signal ids to process")
		return
	}

	h.logger.Debug("[Signaller] querying prices from bothan: %v", nonPendingSignalIDs)
	prices, err := h.bothanClient.QueryPrices(nonPendingSignalIDs)
	if err != nil {
		h.logger.Error("[Signaller] failed to query prices from bothan: %v", err)
		return
	}

	h.logger.Debug("[Signaller] filtering prices")
	submitPrices := h.filterAndPrepareSubmitPrices(prices, nonPendingSignalIDs, now)
	if len(submitPrices) == 0 {
		h.logger.Debug("[Signaller] no prices to submit")
		return
	}

	h.logger.Debug("[Signaller] submitting prices: %v", submitPrices)
	h.submitPrices(submitPrices)
}

func (h *Signaller) submitPrices(prices []types.SignalPrice) {
	for _, p := range prices {
		_, loaded := h.pendingSignalIDs.LoadOrStore(p.SignalID, struct{}{})
		if loaded {
			h.logger.Debug("[Signaller] Attempted to store Signal ID %s which was already pending", p.SignalID)
		}
	}

	h.submitCh <- prices
}

func (h *Signaller) getAllSignalIDs() []string {
	signalIDs := make([]string, 0, len(h.signalIDToFeed))
	for signalID := range h.signalIDToFeed {
		signalIDs = append(signalIDs, signalID)
	}

	return signalIDs
}

func (h *Signaller) getNonPendingSignalIDs() []string {
	signalIDs := h.getAllSignalIDs()

	filtered := make([]string, 0, len(signalIDs))
	for _, signalID := range signalIDs {
		if _, ok := h.pendingSignalIDs.Load(signalID); !ok {
			filtered = append(filtered, signalID)
		}
	}
	return filtered
}

func (h *Signaller) filterAndPrepareSubmitPrices(
	prices []*proto.PriceData,
	signalIDs []string,
	currentTime time.Time,
) []types.SignalPrice {
	pricesMap := sliceToMap(prices, func(price *proto.PriceData) string {
		return price.SignalId
	})

	submitPrices := make([]types.SignalPrice, 0, len(signalIDs))

	for _, signalID := range signalIDs {
		price, ok := pricesMap[signalID]
		if !ok {
			h.logger.Debug("[Signaller] price not found for signal ID: %s", signalID)
			continue
		}

		if !h.isPriceValid(price, currentTime) {
			continue
		}

		submitPrice, err := convertPriceData(price)
		if err != nil {
			h.logger.Debug("[Signaller] failed to parse price data: %v", err)
			continue
		}

		if h.isNonUrgentUnavailablePrices(submitPrice, currentTime.Unix()) {
			h.logger.Debug("[Signaller] non-urgent unavailable price: %v", submitPrice)
			continue
		}

		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices
}

func (h *Signaller) isNonUrgentUnavailablePrices(
	submitPrice types.SignalPrice,
	now int64,
) bool {
	switch submitPrice.PriceStatus {
	case types.PriceStatusUnavailable:
		deadline := h.signalIDToValidatorPrice[submitPrice.SignalID].Timestamp + h.signalIDToFeed[submitPrice.SignalID].Interval
		if now > deadline-FixedIntervalOffset {
			return false
		}
		return true
	default:
		return false
	}
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

	if now.Before(thresholdTime) {
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
