package telemetry

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

var (
	// GroguCollector stores the Cylinder collector instance.
	collector *GroguCollector

	updateSignalPriceTimestampMu = sync.Mutex{}
)

// Metrics is the metrics struct.
type GroguCollector struct {
	Registry                 *prometheus.Registry
	SignalPriceStatus        map[string]feedstypes.SignalPriceStatus
	SignalPriceStatusCount   map[feedstypes.SignalPriceStatus]int
	SignalPriceLatestUpdated map[string]time.Time

	// Updater metrics
	UpdatingRegistryCount      prometheus.Counter // a counter for Bothan registry update.
	UpdateRegistryFailedCount  prometheus.Counter // a counter for an unsuccessful Bothan registry update.
	UpdateRegistrySuccessCount prometheus.Counter // a counter for successful Bothan registry update.

	// Signaler metrics
	ValidatorStatusGauge               prometheus.Gauge    // a gauge for the current validator status.
	ProcessingSignalCount              prometheus.Counter  // a counter for the number of processing signal round.
	ProcessSignalSkippedCount          prometheus.Counter  // a counter for the number of skipped processing signal round.
	ProcessSignalFailedCount           prometheus.Counter  // a counter for the number of failed processing signal round.
	ProcessSignalSuccessCount          prometheus.Counter  // a counter for the number of successful processing signal round.
	QuerySignalPricesDuration          prometheus.Summary  // a summary for the time being consumed for querying signal price from Bothan server.
	NonPendingSignalsGauge             prometheus.Gauge    // a gauge for current signal in the round.
	ConversionErrorSignalsGauge        prometheus.Gauge    // a gauge for the number of signal that failed to convert the result from Bothan server in the round.
	SignalNotFoundGauge                prometheus.Gauge    // a gauge for the number of signal ID that not being found from the list.
	NonUrgentUnavailableSignalIDsGauge prometheus.Gauge    // a gauge for the number of non-urgent signal in the round.
	FilteredSignalingIDsGauge          prometheus.Gauge    // a gauge for the number of signal that should be submitted to BandChain in the round.
	SignalPriceStatusGauge             prometheus.GaugeVec // a gauge for the number of signal per its status (every signals).

	// Submitter metrics
	SubmittingTxCount     prometheus.Counter    // a counter for the number of submitting transaction process.
	SubmitTxFailedCount   prometheus.Counter    // a counter for the number of failed submitting transaction process.
	SubmitTxSuccessCount  prometheus.Counter    // a counter for the number of success submitting transaction process.
	SubmitTxDuration      prometheus.Summary    // a summary for the time being consumed for submitting a transaction to the BandChain.
	WaitingSenderDuration prometheus.Summary    // a summary for the time being consumed for waiting available sender.
	UpdatedSignalInterval prometheus.SummaryVec // a summary for the time interval between the last two updates of the same signal price.
}

// IncrementUpdatingRegistry increments the number of sending a Bothan's registry update request.
func IncrementUpdatingRegistry() {
	if collector == nil {
		return
	}

	collector.UpdatingRegistryCount.Inc()
}

// IncrementUpdateRegistryFailed increments the number of failed Bothan's registry update request.
func IncrementUpdateRegistryFailed() {
	if collector == nil {
		return
	}

	collector.UpdateRegistryFailedCount.Inc()
}

// IncrementUpdateRegistrySuccess increments the number of successful Bothan's registry update request.
func IncrementUpdateRegistrySuccess() {
	if collector == nil {
		return
	}

	collector.UpdateRegistrySuccessCount.Inc()
}

// SetValidatorStatus sets the validator status.
func SetValidatorStatus(status bool) {
	if collector == nil {
		return
	}

	statusValue := 0.0
	if status {
		statusValue = 1.0
	}

	collector.ValidatorStatusGauge.Set(statusValue)
}

// IncrementProcessingSignal increments the number of processing signal round.
func IncrementProcessingSignal() {
	if collector == nil {
		return
	}

	collector.ProcessingSignalCount.Inc()
}

// IncrementProcessSignalSkipped increments the number of processing signal round that being skipped.
func IncrementProcessSignalSkipped() {
	if collector == nil {
		return
	}

	collector.ProcessSignalSkippedCount.Inc()
}

// IncrementProcessSignalFailed increments the number of failed processing signal round.
func IncrementProcessSignalFailed() {
	if collector == nil {
		return
	}

	collector.ProcessSignalFailedCount.Inc()
}

// IncrementProcessSignalSuccess increments the number of successful processing signal round.
func IncrementProcessSignalSuccess() {
	if collector == nil {
		return
	}

	collector.ProcessSignalSuccessCount.Inc()
}

// ObserveQuerySignalPricesDuration observes the time being consumed for querying signal price
// from Bothan server.
func ObserveQuerySignalPricesDuration(duration float64) {
	if collector == nil {
		return
	}

	collector.QuerySignalPricesDuration.Observe(duration)
}

// SetNonPendingSignals sets the number of non-pending signal in the round.
func SetNonPendingSignals(count int) {
	if collector == nil {
		return
	}

	collector.NonPendingSignalsGauge.Set(float64(count))
}

// SetConversionErrorSignals sets the number of signal that failed to convert the result
// from Bothan server in the round.
func SetConversionErrorSignals(count int) {
	if collector == nil {
		return
	}

	collector.ConversionErrorSignalsGauge.Set(float64(count))
}

// SetSignalNotFound sets the number of signal ID that not being found from the list.
func SetSignalNotFound(count int) {
	if collector == nil {
		return
	}

	collector.SignalNotFoundGauge.Set(float64(count))
}

// SetNonUrgentUnavailablePriceSignals sets the number of non-urgent signal in the round.
func SetNonUrgentUnavailablePriceSignals(count int) {
	if collector == nil {
		return
	}

	collector.NonUrgentUnavailableSignalIDsGauge.Set(float64(count))
}

// SetFilteredSignalIDs sets the number of signal that should be submitted to BandChain in the round.
func SetFilteredSignalIDs(count int) {
	if collector == nil {
		return
	}

	collector.FilteredSignalingIDsGauge.Set(float64(count))
}

// SetSignalPriceStatuses sets the current number of signal price status.
func SetSignalPriceStatuses(newSignalPrices []feedstypes.SignalPrice) {
	if collector == nil {
		return
	}

	for _, signal := range newSignalPrices {
		if oldStatus, ok := collector.SignalPriceStatus[signal.SignalID]; ok {
			collector.SignalPriceStatusCount[oldStatus]--
		}

		collector.SignalPriceStatus[signal.SignalID] = signal.Status
		collector.SignalPriceStatusCount[signal.Status]++
	}

	// Update signal price status gauge
	for status, statusName := range feedstypes.SignalPriceStatus_name {
		collector.SignalPriceStatusGauge.WithLabelValues(statusName).
			Set(float64(collector.SignalPriceStatusCount[feedstypes.SignalPriceStatus(status)]))
	}
}

// IncrementSubmittingTx increments the number of submitting transaction process.
func IncrementSubmittingTx() {
	if collector == nil {
		return
	}

	collector.SubmittingTxCount.Inc()
}

// IncrementSubmitTxFailed increments the number of failed submitting transaction process.
func IncrementSubmitTxFailed() {
	if collector == nil {
		return
	}

	collector.SubmitTxFailedCount.Inc()
}

// IncrementSubmitTxSuccess increments the number of success submitting transaction process.
func IncrementSubmitTxSuccess() {
	if collector == nil {
		return
	}

	collector.SubmitTxSuccessCount.Inc()
}

// ObserveSubmitTxDuration observes the time being consumed for submitting a transaction to the BandChain.
func ObserveSubmitTxDuration(duration float64) {
	if collector == nil {
		return
	}

	collector.SubmitTxDuration.Observe(duration)
}

// ObserveWaitingSenderDuration observes the time being consumed for waiting available sender.
func ObserveWaitingSenderDuration(duration float64) {
	if collector == nil {
		return
	}

	collector.WaitingSenderDuration.Observe(duration)
}

// ObserveSignalPriceUpdateInterval observes the time interval between the last two updates of the same signal price.
func ObserveSignalPriceUpdateInterval(signalPrices []feedstypes.SignalPrice) {
	if collector == nil {
		return
	}

	updateSignalPriceTimestampMu.Lock()
	defer updateSignalPriceTimestampMu.Unlock()

	now := time.Now()
	for _, signal := range signalPrices {
		if lastUpdated, ok := collector.SignalPriceLatestUpdated[signal.SignalID]; ok {
			collector.UpdatedSignalInterval.WithLabelValues(signal.SignalID).
				Observe(now.Sub(lastUpdated).Seconds())
		}

		collector.SignalPriceLatestUpdated[signal.SignalID] = now
	}
}

// NewGroguCollector creates a new cylinder collector instance.
func NewGroguCollector(labels prometheus.Labels) *GroguCollector {
	registry := prometheus.NewRegistry()
	registerer := promauto.With(registry)
	// metrics for updater
	updatingRegistryCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_update_registry_count",
		Help:        "number of times the registry is updated since last grogu restart",
		ConstLabels: labels,
	})
	updateRegistryFailedCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_update_registry_failed_count",
		Help:        "number of times the registry fail to update since last grogu restart",
		ConstLabels: labels,
	})
	updateRegistrySuccessCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_update_registry_success_count",
		Help:        "number of times the registry successfully update since last grogu restart",
		ConstLabels: labels,
	})

	// metrics for signaler
	validatorStatusGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_validator_status",
		Help:        "validator status (1 = active, 0 = inactive)",
		ConstLabels: labels,
	})
	processingSignalCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_processing_signal_count",
		Help:        "number of times the signaler processes signal prices",
		ConstLabels: labels,
	})
	processSignalSkippedCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_process_signal_skipped_count",
		Help:        "number of times the signaler's process is skipped",
		ConstLabels: labels,
	})
	processSignalFailedCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_process_signal_failed_count",
		Help:        "number of times the signaler failed to process signal prices",
		ConstLabels: labels,
	})
	processSignalSuccessCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_process_signal_success_count",
		Help:        "number of times the signaler successfully process signal prices",
		ConstLabels: labels,
	})
	querySignalPricesDuration := registerer.NewSummary(prometheus.SummaryOpts{
		Name: "grogu_query_signal_prices_duration",
		Help: "time being consumed for querying signal prices from Bothan service",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
		ConstLabels: labels,
	})
	nonPendingSignalsGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_non_pending_signal_ids",
		Help:        "number of non-pending signal IDs in the signaling round",
		ConstLabels: labels,
	})
	conversionErrorSignalsGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_conversion_error_signal_ids",
		Help:        "number of signal IDs that failed to convert to signal prices in the signaling round",
		ConstLabels: labels,
	})
	signalNotFoundGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_signal_not_found",
		Help:        "number of signal IDs that aren't found from the price list",
		ConstLabels: labels,
	})
	nonUrgentUnavailableSignalIDsGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_non_urgent_unavailable_signal_ids",
		Help:        "number of signal IDs that the signal price whose status is unavailable and isn't urgent in the signaling round",
		ConstLabels: labels,
	})
	filteredSignalingIDsGauge := registerer.NewGauge(prometheus.GaugeOpts{
		Name:        "grogu_filtered_signal_ids",
		Help:        "number of signal IDs that is allowed to submit to the BandChain in the signaling round",
		ConstLabels: labels,
	})
	signalPriceStatusGauge := *registerer.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "grogu_signal_price_status",
		Help:        "number of signal prices with specific status",
		ConstLabels: labels,
	}, []string{"signal_price_status"})

	// metrics for submitter
	submittingTxCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_submitting_tx_count",
		Help:        "number of times the submitter submits transactions",
		ConstLabels: labels,
	})
	submitTxFailedCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_submit_tx_failed_count",
		Help:        "number of times the submitter fail to submit transactions",
		ConstLabels: labels,
	})
	submitTxSuccessCount := registerer.NewCounter(prometheus.CounterOpts{
		Name:        "grogu_submit_tx_success_count",
		Help:        "number of times the submitter successfully submits transactions",
		ConstLabels: labels,
	})
	submitTxDuration := registerer.NewSummary(prometheus.SummaryOpts{
		Name: "grogu_submit_tx_duration",
		Help: "time being consumed for submitting a transaction to the BandChain",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
		ConstLabels: labels,
	})
	waitingSenderDuration := registerer.NewSummary(prometheus.SummaryOpts{
		Name: "grogu_waiting_sender_duration",
		Help: "time being consumed for waiting the free sender for sending a transaction",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
		ConstLabels: labels,
	})
	updatedSignalInterval := *registerer.NewSummaryVec(prometheus.SummaryOpts{
		Name: "grogu_updated_signal_interval",
		Help: "time interval between the last two updates of the same signal price",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
		ConstLabels: labels,
	}, []string{"signal_id"})

	return &GroguCollector{
		Registry:                           registry,
		SignalPriceStatus:                  make(map[string]feedstypes.SignalPriceStatus),
		SignalPriceStatusCount:             make(map[feedstypes.SignalPriceStatus]int),
		SignalPriceLatestUpdated:           make(map[string]time.Time),
		UpdatingRegistryCount:              updatingRegistryCount,
		UpdateRegistryFailedCount:          updateRegistryFailedCount,
		UpdateRegistrySuccessCount:         updateRegistrySuccessCount,
		ValidatorStatusGauge:               validatorStatusGauge,
		ProcessingSignalCount:              processingSignalCount,
		ProcessSignalSkippedCount:          processSignalSkippedCount,
		ProcessSignalFailedCount:           processSignalFailedCount,
		ProcessSignalSuccessCount:          processSignalSuccessCount,
		QuerySignalPricesDuration:          querySignalPricesDuration,
		NonPendingSignalsGauge:             nonPendingSignalsGauge,
		ConversionErrorSignalsGauge:        conversionErrorSignalsGauge,
		SignalNotFoundGauge:                signalNotFoundGauge,
		NonUrgentUnavailableSignalIDsGauge: nonUrgentUnavailableSignalIDsGauge,
		FilteredSignalingIDsGauge:          filteredSignalingIDsGauge,
		SignalPriceStatusGauge:             signalPriceStatusGauge,
		SubmittingTxCount:                  submittingTxCount,
		SubmitTxFailedCount:                submitTxFailedCount,
		SubmitTxSuccessCount:               submitTxSuccessCount,
		SubmitTxDuration:                   submitTxDuration,
		WaitingSenderDuration:              waitingSenderDuration,
		UpdatedSignalInterval:              updatedSignalInterval,
	}
}

// Describe sends the descriptors of each metric to the provided channel.
func (c GroguCollector) Describe(ch chan<- *prometheus.Desc) {
	// description for updater
	ch <- c.UpdatingRegistryCount.Desc()
	ch <- c.UpdateRegistryFailedCount.Desc()
	ch <- c.UpdateRegistrySuccessCount.Desc()

	// description for signaler
	ch <- c.ValidatorStatusGauge.Desc()
	ch <- c.ProcessingSignalCount.Desc()
	ch <- c.ProcessSignalSkippedCount.Desc()
	ch <- c.ProcessSignalFailedCount.Desc()
	ch <- c.ProcessSignalSuccessCount.Desc()
	ch <- c.QuerySignalPricesDuration.Desc()
	ch <- c.NonPendingSignalsGauge.Desc()
	ch <- c.ConversionErrorSignalsGauge.Desc()
	ch <- c.SignalNotFoundGauge.Desc()
	ch <- c.NonUrgentUnavailableSignalIDsGauge.Desc()
	ch <- c.FilteredSignalingIDsGauge.Desc()
	c.SignalPriceStatusGauge.Describe(ch)

	// description for submitter
	ch <- c.SubmittingTxCount.Desc()
	ch <- c.SubmitTxFailedCount.Desc()
	ch <- c.SubmitTxSuccessCount.Desc()
	ch <- c.SubmitTxDuration.Desc()
	ch <- c.WaitingSenderDuration.Desc()
	c.UpdatedSignalInterval.Describe(ch)
}

// Collect sends the metric values for each metric related to the Cylinder collector to the provided channel.
func (c GroguCollector) Collect(ch chan<- prometheus.Metric) {
	// collector for updater
	ch <- c.UpdatingRegistryCount
	ch <- c.UpdateRegistryFailedCount
	ch <- c.UpdateRegistrySuccessCount

	// collector for signaler
	ch <- c.ValidatorStatusGauge
	ch <- c.ProcessingSignalCount
	ch <- c.ProcessSignalSkippedCount
	ch <- c.ProcessSignalFailedCount
	ch <- c.ProcessSignalSuccessCount
	ch <- c.QuerySignalPricesDuration
	ch <- c.NonPendingSignalsGauge
	ch <- c.ConversionErrorSignalsGauge
	ch <- c.SignalNotFoundGauge
	ch <- c.NonUrgentUnavailableSignalIDsGauge
	ch <- c.FilteredSignalingIDsGauge
	c.SignalPriceStatusGauge.Collect(ch)

	// description for submitter
	ch <- c.SubmittingTxCount
	ch <- c.SubmitTxFailedCount
	ch <- c.SubmitTxSuccessCount
	ch <- c.SubmitTxDuration
	ch <- c.WaitingSenderDuration
	c.UpdatedSignalInterval.Collect(ch)
}
