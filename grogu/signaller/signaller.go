package signaller

import (
	"sync"
	"time"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/price"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	FixedIntervalOffset int64 = 10
	TimeBuffer          int64 = 3
)

type Signaller struct {
	feedQuerier  FeedQuerier
	bothanClient BothanClient
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
	feedQuerier FeedQuerier,
	bothanClient BothanClient,
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

func (s *Signaller) Start() {
	for {
		time.Sleep(s.interval)

		resp, err := s.feedQuerier.QueryValidValidator(s.valAddress)
		if err != nil {
			s.logger.Error("[Signaller] failed to query valid validator: %v", err)
			continue
		}

		if !resp.Valid {
			s.logger.Info("[Signaller] validator is not required to feed prices")
			continue
		}

		if !s.updateInternalVariables() {
			s.logger.Error("[Signaller] failed to update internal variables")
			continue
		}

		s.execute()
	}
}

func (s *Signaller) updateInternalVariables() bool {
	resultCh := make(chan bool, 3)
	var wg sync.WaitGroup

	updater := func(f func() bool) {
		defer wg.Done()
		resultCh <- f()
	}

	wg.Add(3)
	go updater(s.updateParams)
	go updater(s.updateFeedMap)
	go updater(s.updateValidatorPriceMap)
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

func (s *Signaller) updateParams() bool {
	resp, err := s.feedQuerier.QueryParams()
	if err != nil {
		s.logger.Error("[Signaller] failed to query params: %v", err)
		return false
	}

	s.params = &resp.Params
	return true
}

func (s *Signaller) updateFeedMap() bool {
	resp, err := s.feedQuerier.QueryCurrentFeeds()
	if err != nil {
		s.logger.Error("[Signaller] failed to query supported feeds: %v", err)
		return false
	}

	s.signalIDToFeed = sliceToMap(resp.CurrentFeeds.Feeds, func(feed types.FeedWithDeviation) string {
		return feed.SignalID
	})

	return true
}

func (s *Signaller) updateValidatorPriceMap() bool {
	resp, err := s.feedQuerier.QueryValidatorPrices(s.valAddress)
	if err != nil {
		s.logger.Error("[Signaller] failed to query validator prices: %v", err)
		return false
	}

	s.signalIDToValidatorPrice = sliceToMap(resp.ValidatorPrices, func(valPrice types.ValidatorPrice) string {
		return valPrice.SignalID
	})

	return true
}

func (s *Signaller) execute() {
	now := time.Now()

	s.logger.Debug("[Signaller] starting signal process")

	s.logger.Debug("[Signaller] getting non-pending signal ids")
	nonPendingSignalIDs := s.getNonPendingSignalIDs()
	if len(nonPendingSignalIDs) == 0 {
		s.logger.Debug("[Signaller] no signal ids to process")
		return
	}

	s.logger.Debug("[Signaller] querying prices from bothan: %v", nonPendingSignalIDs)
	prices, err := s.bothanClient.GetPrices(nonPendingSignalIDs)
	if err != nil {
		s.logger.Error("[Signaller] failed to query prices from bothan: %v", err)
		return
	}

	s.logger.Debug("[Signaller] filtering prices")
	submitPrices := s.filterAndPrepareSubmitPrices(prices, nonPendingSignalIDs, now)
	if len(submitPrices) == 0 {
		s.logger.Debug("[Signaller] no prices to submit")
		return
	}

	s.logger.Debug("[Signaller] submitting prices: %v", submitPrices)
	s.submitPrices(submitPrices)
}

func (s *Signaller) submitPrices(prices []types.SignalPrice) {
	for _, p := range prices {
		_, loaded := s.pendingSignalIDs.LoadOrStore(p.SignalID, struct{}{})
		if loaded {
			s.logger.Debug("[Signaller] Attempted to store Signal ID %s which was already pending", p.SignalID)
		}
	}

	s.submitCh <- prices
}

func (s *Signaller) getAllSignalIDs() []string {
	signalIDs := make([]string, 0, len(s.signalIDToFeed))
	for signalID := range s.signalIDToFeed {
		signalIDs = append(signalIDs, signalID)
	}

	return signalIDs
}

func (s *Signaller) getNonPendingSignalIDs() []string {
	signalIDs := s.getAllSignalIDs()

	filtered := make([]string, 0, len(signalIDs))
	for _, signalID := range signalIDs {
		if _, ok := s.pendingSignalIDs.Load(signalID); !ok {
			filtered = append(filtered, signalID)
		}
	}
	return filtered
}

func (s *Signaller) filterAndPrepareSubmitPrices(
	prices []*bothan.Price,
	signalIDs []string,
	currentTime time.Time,
) []types.SignalPrice {
	pricesMap := sliceToMap(prices, func(price *bothan.Price) string {
		return price.SignalId
	})

	submitPrices := make([]types.SignalPrice, 0, len(signalIDs))

	for _, signalID := range signalIDs {
		price, ok := pricesMap[signalID]
		if !ok {
			s.logger.Debug("[Signaller] price not found for signal ID: %s", signalID)
			continue
		}

		if !s.isPriceValid(price, currentTime) {
			continue
		}

		submitPrice, err := convertPriceData(price)
		if err != nil {
			s.logger.Debug("[Signaller] failed to parse price data: %v", err)
			continue
		}

		if s.isNonUrgentUnavailablePrices(submitPrice, currentTime.Unix()) {
			s.logger.Debug("[Signaller] non-urgent unavailable price: %v", submitPrice)
			continue
		}

		submitPrices = append(submitPrices, submitPrice)
	}

	return submitPrices
}

func (s *Signaller) isNonUrgentUnavailablePrices(
	submitPrice types.SignalPrice,
	now int64,
) bool {
	switch submitPrice.PriceStatus {
	case types.PriceStatusUnavailable:
		deadline := s.signalIDToValidatorPrice[submitPrice.SignalID].Timestamp + s.signalIDToFeed[submitPrice.SignalID].Interval
		if now > deadline-FixedIntervalOffset {
			return false
		}
		return true
	default:
		return false
	}
}

func (s *Signaller) isPriceValid(
	price *bothan.Price,
	now time.Time,
) bool {
	// Check if the price is supported and required to be submitted
	feed, ok := s.signalIDToFeed[price.SignalId]
	if !ok {
		return false
	}

	// Get the last price submitted by the validator, if it doesn't exist, it is valid to be sent
	valPrice, ok := s.signalIDToValidatorPrice[price.SignalId]
	if !ok {
		return true
	}

	// If the last price exists, check if the price can be updated
	if s.shouldUpdatePrice(feed, valPrice, price.Price, now) {
		return true
	}

	return false
}

func (s *Signaller) shouldUpdatePrice(
	feed types.FeedWithDeviation,
	valPrice types.ValidatorPrice,
	newPrice uint64,
	now time.Time,
) bool {
	// thresholdTime is the time when the price can be updated.
	// add TimeBuffer to make sure the thresholdTime is not too early.
	thresholdTime := time.Unix(valPrice.Timestamp+s.params.CooldownTime+TimeBuffer, 0)

	if now.Before(thresholdTime) {
		return false
	}

	// Check if the price is past the assigned time, if it is, add it to the list of prices to update
	assignedTime := calculateAssignedTime(
		s.valAddress,
		feed.Interval,
		valPrice.Timestamp,
		s.distributionOffsetPercentage,
		s.distributionStartPercentage,
	)

	if !now.Before(assignedTime) {
		return true
	}

	// Check if the price is deviated from the last submission, if it is, add it to the list of prices to update
	if isDeviated(feed.DeviationBasisPoint, valPrice.Price, newPrice) {
		return true
	}

	return false
}
