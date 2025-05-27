package signaller

import (
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/bothan/v1"

	"github.com/bandprotocol/chain/v3/grogu/submitter"
	"github.com/bandprotocol/chain/v3/grogu/telemetry"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

const (
	// FixedIntervalOffset is the offset to be added to the interval to get the deadline
	FixedIntervalOffset int64 = 15
)

type Signaller struct {
	feedQuerier  FeedQuerier
	nodeQuerier  NodeQuerier
	bothanClient BothanClient
	// How often to check for signal changes
	interval         time.Duration
	submitCh         chan<- submitter.SignalPriceSubmission
	logger           *logger.Logger
	valAddress       sdk.ValAddress
	pendingSignalIDs *sync.Map

	distributionStartPercentage  uint64
	distributionOffsetPercentage uint64

	signalIDToFeed           map[string]types.FeedWithDeviation
	signalIDToValidatorPrice map[string]types.ValidatorPrice
	params                   *types.Params
	currentBlockTime         time.Time
}

func New(
	feedQuerier FeedQuerier,
	nodeQuerier NodeQuerier,
	bothanClient BothanClient,
	interval time.Duration,
	submitCh chan<- submitter.SignalPriceSubmission,
	logger *logger.Logger,
	valAddress sdk.ValAddress,
	pendingSignalIDs *sync.Map,
	distributionStartPercentage uint64,
	distributionOffsetPercentage uint64,
) *Signaller {
	return &Signaller{
		feedQuerier:                  feedQuerier,
		nodeQuerier:                  nodeQuerier,
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
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := s.feedQuerier.QueryValidValidator(s.valAddress)
		if err != nil {
			s.logger.Error("[Signaller] failed to query valid validator: %v", err)
			continue
		}

		telemetry.SetValidatorStatus(resp.Valid)

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
	updaters := []func() bool{
		s.updateParams,
		s.updateFeedMap,
		s.updateValidatorPriceMap,
		s.updateBlockTime,
	}

	resultCh := make(chan bool, len(updaters))
	var wg sync.WaitGroup

	for _, updater := range updaters {
		wg.Add(1)
		go func(f func() bool) {
			defer wg.Done()
			resultCh <- f()
		}(updater)
	}

	wg.Wait()
	close(resultCh)

	for result := range resultCh {
		if !result {
			return false
		}
	}

	return true
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

func (s *Signaller) updateBlockTime() bool {
	resp, err := s.nodeQuerier.QueryStatus()
	if err != nil {
		s.logger.Error("[Signaller] failed to query latest block: %v", err)
		return false
	}

	s.currentBlockTime = *resp.Timestamp

	return true
}

func (s *Signaller) execute() {
	telemetry.IncrementProcessingSignal()
	s.logger.Debug("[Signaller] starting signal process")

	s.logger.Debug("[Signaller] getting non-pending signal ids")
	nonPendingSignalIDs := s.getNonPendingSignalIDs()
	telemetry.SetNonPendingSignals(len(nonPendingSignalIDs))

	if len(nonPendingSignalIDs) == 0 {
		telemetry.IncrementProcessSignalSkipped()
		s.logger.Debug("[Signaller] no signal ids to process")
		return
	}

	s.logger.Debug("[Signaller] querying prices from bothan: %v", nonPendingSignalIDs)

	since := time.Now()
	res, err := s.bothanClient.GetPrices(nonPendingSignalIDs)
	if err != nil {
		telemetry.IncrementProcessSignalFailed()
		s.logger.Error("[Signaller] failed to query prices from bothan: %v", err)
		return
	}
	telemetry.ObserveQuerySignalPricesDuration(time.Since(since).Seconds())

	prices, uuid := res.Prices, res.Uuid

	s.logger.Debug("[Signaller] filtering prices")

	signalPrices := s.filterAndPrepareSignalPrices(prices, nonPendingSignalIDs)
	telemetry.SetFilteredSignalIDs(len(signalPrices))
	if len(signalPrices) == 0 {
		telemetry.IncrementProcessSignalSkipped()
		s.logger.Debug("[Signaller] no prices to submit")
		return
	}

	s.logger.Debug("[Signaller] submitting prices: %v", signalPrices)
	s.submitPrices(signalPrices, uuid)

	telemetry.SetSignalPriceStatuses(signalPrices)
	telemetry.IncrementProcessSignalSuccess()
}

func (s *Signaller) submitPrices(prices []types.SignalPrice, uuid string) {
	for _, p := range prices {
		_, loaded := s.pendingSignalIDs.LoadOrStore(p.SignalID, struct{}{})
		if loaded {
			s.logger.Debug("[Signaller] Attempted to store Signal ID %s which was already pending", p.SignalID)
		}
	}

	signalPriceSubmission := submitter.SignalPriceSubmission{
		SignalPrices: prices,
		UUID:         uuid,
	}

	s.submitCh <- signalPriceSubmission
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

func (s *Signaller) filterAndPrepareSignalPrices(
	prices []*bothan.Price,
	signalIDs []string,
) []types.SignalPrice {
	pricesMap := sliceToMap(prices, func(price *bothan.Price) string {
		return price.SignalId
	})

	signalPrices := make([]types.SignalPrice, 0, len(signalIDs))
	conversionErrorCnt := 0
	signalNotFoundCnt := 0
	nonUrgentUnavailablePriceCnt := 0

	for _, signalID := range signalIDs {
		price, ok := pricesMap[signalID]
		if !ok {
			signalNotFoundCnt++
			s.logger.Debug("[Signaller] price not found for signal ID: %s", signalID)
			continue
		}

		signalPrice, err := convertPriceData(price)
		if err != nil {
			conversionErrorCnt++
			s.logger.Debug("[Signaller] failed to parse price data: %v", err)
			continue
		}

		if !s.isPriceValid(signalPrice) {
			continue
		}

		if s.isNonUrgentUnavailablePrices(signalPrice) {
			nonUrgentUnavailablePriceCnt++
			s.logger.Debug("[Signaller] non-urgent unavailable price: %v", signalPrice)
			continue
		}

		signalPrices = append(signalPrices, signalPrice)
	}

	telemetry.SetConversionErrorSignals(conversionErrorCnt)
	telemetry.SetSignalNotFound(signalNotFoundCnt)
	telemetry.SetNonUrgentUnavailablePriceSignals(nonUrgentUnavailablePriceCnt)

	return signalPrices
}

func (s *Signaller) isNonUrgentUnavailablePrices(
	signalPrice types.SignalPrice,
) bool {
	switch signalPrice.Status {
	case types.SIGNAL_PRICE_STATUS_UNAVAILABLE:
		deadline := s.signalIDToValidatorPrice[signalPrice.SignalID].Timestamp + s.signalIDToFeed[signalPrice.SignalID].Interval
		if s.currentBlockTime.Unix() >= deadline-FixedIntervalOffset {
			return false
		}
		return true
	default:
		return false
	}
}

func (s *Signaller) isPriceValid(
	newPrice types.SignalPrice,
) bool {
	// Check if the price is supported and required to be submitted
	feed, ok := s.signalIDToFeed[newPrice.SignalID]
	if !ok {
		return false
	}

	// Get the last price submitted by the validator, if it doesn't exist, it is valid to be sent
	oldPrice, ok := s.signalIDToValidatorPrice[newPrice.SignalID]
	if !ok {
		return true
	}

	// If the last price exists, check if the price can be updated
	return s.shouldUpdatePrice(feed, oldPrice, newPrice)
}

func (s *Signaller) shouldUpdatePrice(
	feed types.FeedWithDeviation,
	oldPrice types.ValidatorPrice,
	newPrice types.SignalPrice,
) bool {
	// thresholdTime is the time when the price can be updated.
	thresholdTime := oldPrice.Timestamp + s.params.CooldownTime

	if s.currentBlockTime.Unix() < thresholdTime {
		return false
	}

	// Check if the price is past the assigned time, if it is, add it to the list of prices to update
	assignedTime := calculateAssignedTime(
		s.valAddress,
		feed.Interval,
		oldPrice.Timestamp,
		s.distributionStartPercentage,
		s.distributionOffsetPercentage,
	)

	if !s.currentBlockTime.Before(assignedTime) {
		return true
	}

	if oldPrice.SignalPriceStatus != newPrice.Status {
		return true
	}

	// Check if the price is deviated from the last submission, if it is, add it to the list of prices to update
	return isDeviated(feed.DeviationBasisPoint, oldPrice.Price, newPrice.Price)
}
