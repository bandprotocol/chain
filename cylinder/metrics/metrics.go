package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	cylinderctx "github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// metrics stores the Prometheus metrics instance.
var metrics *PrometheusMetrics

// globalTelemetryEnabled indicates whether telemetry is enabled globally.
// It is set on initialization and does not change for the lifetime of the program.
var globalTelemetryEnabled bool

type PrometheusMetrics struct {
	// process group metrics
	ProcessRound1SuccessCount  *prometheus.CounterVec
	ProcessRound1FailureCount  *prometheus.CounterVec
	ProcessRound1Time          *prometheus.SummaryVec
	ProcessRound2SuccessCount  *prometheus.CounterVec
	ProcessRound2FailureCount  *prometheus.CounterVec
	ProcessRound2Time          *prometheus.SummaryVec
	ProcessRound3ConfirmCount  *prometheus.CounterVec
	ProcessRound3ComplainCount *prometheus.CounterVec
	ProcessRound3FailureCount  *prometheus.CounterVec
	ProcessRound3Time          *prometheus.SummaryVec
	DKGLeftGauge               prometheus.Gauge
	GroupCount                 prometheus.Counter

	// Member metrics
	MemberStatusGauge *prometheus.GaugeVec

	// DE metrics
	OnChinDELeftGauge   prometheus.Gauge
	OffChainDELeftGauge prometheus.Gauge

	// signing metrics
	IncomingSigningCount       *prometheus.CounterVec
	ProcessSigningSuccessCount *prometheus.CounterVec
	ProcessSigningFailureCount *prometheus.CounterVec
	ProcessSigningTime         *prometheus.SummaryVec

	// Submitter metrics
	WaitingSenderTime    prometheus.Summary
	SubmittingTxCount    prometheus.Counter
	SubmitTxSuccessCount prometheus.Counter
	SubmitTxFailedCount  prometheus.Counter
	SubmitTxTime         prometheus.Summary
}

func updateMetrics(updateFn func()) {
	if globalTelemetryEnabled {
		updateFn()
	}
}

// IncProcessRound1SuccessCount increments the count of successful round 1 executions for a specific group.
func IncProcessRound1SuccessCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound1SuccessCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessRound1FailureCount increments the count of failed round 1 executions for a specific group.
func IncProcessRound1FailureCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound1FailureCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// ObserveProcessRound1Time observes the time taken to process round 1 for a specific group.
func ObserveProcessRound1Time(groupID uint64, duration float64) {
	updateMetrics(func() {
		metrics.ProcessRound1Time.WithLabelValues(fmt.Sprintf("%d", groupID)).Observe(duration)
	})
}

// IncProcessRound2SuccessCount increments the count of successful round 2 executions for a specific group.
func IncProcessRound2SuccessCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound2SuccessCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessRound2FailureCount increments the count of failed round 2 executions for a specific group.
func IncProcessRound2FailureCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound2FailureCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// ObserveProcessRound2Time observes the time taken to process round 2 for a specific group.
func ObserveProcessRound2Time(groupID uint64, duration float64) {
	updateMetrics(func() {
		metrics.ProcessRound2Time.WithLabelValues(fmt.Sprintf("%d", groupID)).Observe(duration)
	})
}

// IncProcessRound3ConfirmCount increments the count of successful round 3 confirmations for a specific group.
func IncProcessRound3ConfirmCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound3ConfirmCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessRound3ComplainCount increments the count of round 3 complaints for a specific group.
func IncProcessRound3ComplainCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound3ComplainCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessRound3FailureCount increments the count of failed round 3 executions for a specific group.
func IncProcessRound3FailureCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessRound3FailureCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// ObserveProcessRound3Time observes the time taken to process round 3 for a specific group.
func ObserveProcessRound3Time(groupID uint64, duration float64) {
	updateMetrics(func() {
		metrics.ProcessRound3Time.WithLabelValues(fmt.Sprintf("%d", groupID)).Observe(duration)
	})
}

// IncDKGLeftGauge increments the value of the DKG left gauge.
func IncDKGLeftGauge() {
	updateMetrics(func() {
		metrics.DKGLeftGauge.Inc()
	})
}

// AddDKGLeftGauge adds the value to the DKG left gauge.
func AddDKGLeftGauge(n float64) {
	updateMetrics(func() {
		metrics.DKGLeftGauge.Add(n)
	})
}

// DecDKGLeftGauge decrements the value of the DKG left gauge.
func DecDKGLeftGauge() {
	updateMetrics(func() {
		metrics.DKGLeftGauge.Dec()
	})
}

// IncGroupCount increments the count of groups.
func IncGroupCount() {
	updateMetrics(func() {
		metrics.GroupCount.Inc()
	})
}

// AddGroupCount adds the number of groups.
func AddGroupCount(n float64) {
	updateMetrics(func() {
		metrics.GroupCount.Add(n)
	})
}

// SetMemberStatus sets the status of a member.
func SetMemberStatus(groupID uint64, status bool) {
	statusValue := 0.0
	if status {
		statusValue = 1.0
	}

	metrics.MemberStatusGauge.WithLabelValues(fmt.Sprintf("%d", groupID)).Set(statusValue)
}

// SetOnChainDELeftGauge sets the value of the on-chain DE left gauge.
func SetOnChainDELeftGauge(value float64) {
	updateMetrics(func() {
		metrics.OnChinDELeftGauge.Set(value)
	})
}

// IncOffChainDELeftGauge increments the value of the off-chain DE left gauge.
func IncOffChainDELeftGauge() {
	updateMetrics(func() {
		metrics.OffChainDELeftGauge.Inc()
	})
}

// AddOffChainDELeftGauge adds the value to the off-chain DE left gauge.
func AddOffChainDELeftGauge(n float64) {
	updateMetrics(func() {
		metrics.OffChainDELeftGauge.Add(n)
	})
}

// DecOffChainDELeftGauge decrements the value of the off-chain DE left gauge.
func DecOffChainDELeftGauge() {
	updateMetrics(func() {
		metrics.OffChainDELeftGauge.Dec()
	})
}

// IncIncomingSigningCount increments the count of incoming signing requests for a specific group.
func IncIncomingSigningCount(groupID uint64) {
	updateMetrics(func() {
		metrics.IncomingSigningCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessSigningSuccessCount increments the count of successful process signing for a specific group.
func IncProcessSigningSuccessCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessSigningSuccessCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// IncProcessSigningFailureCount increments the count of failed process signing for a specific group.
func IncProcessSigningFailureCount(groupID uint64) {
	updateMetrics(func() {
		metrics.ProcessSigningFailureCount.WithLabelValues(fmt.Sprintf("%d", groupID)).Inc()
	})
}

// ObserveProcessSigningTime observes the time taken to process signing for a specific group.
func ObserveProcessSigningTime(groupID uint64, duration float64) {
	updateMetrics(func() {
		metrics.ProcessSigningTime.WithLabelValues(fmt.Sprintf("%d", groupID)).Observe(duration)
	})
}

// ObserveWaitingSenderTime observes the time taken to wait for a free key.
func ObserveWaitingSenderTime(duration float64) {
	updateMetrics(func() {
		metrics.WaitingSenderTime.Observe(duration)
	})
}

// AddSubmittingTxCount adds the number of submitting transactions.
func AddSubmittingTxCount(n float64) {
	updateMetrics(func() {
		metrics.SubmittingTxCount.Add(n)
	})
}

// IncSubmitTxSuccessCount increments the count of successful submit transactions.
func IncSubmitTxSuccessCount() {
	updateMetrics(func() {
		metrics.SubmitTxSuccessCount.Inc()
	})
}

// IncSubmitTxFailedCount increments the count of failed submit transactions.
func IncSubmitTxFailedCount() {
	updateMetrics(func() {
		metrics.SubmitTxFailedCount.Inc()
	})
}

// ObserveSubmitTxTime observes the time taken to submit transactions.
func ObserveSubmitTxTime(duration float64) {
	updateMetrics(func() {
		metrics.SubmitTxTime.Observe(duration)
	})
}

func InitPrometheusMetrics(labels prometheus.Labels) {
	roundLabels := []string{"group_id"}
	memberLabels := []string{"group_id"}
	signingLabels := []string{"group_id"}

	metrics = &PrometheusMetrics{
		ProcessRound1SuccessCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_round1_success_count",
			Help:        "Number of successful process round 1",
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound1FailureCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_round1_failure_count",
			Help:        "Number of failed process round 1",
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound1Time: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "cylinder_process_round1_time",
			Help: "Time taken to process round 1",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound2SuccessCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_round2_success_count",
			Help:        "Number of successful process round 2",
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound2FailureCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_round2_failure_count",
			Help:        "Number of failed process round 2",
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound2Time: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "cylinder_process_round2_time",
			Help: "Time taken to process round 2",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound3ConfirmCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_round3_confirm_count",
			Help:        "Number of successful process round 3 confirm",
			ConstLabels: labels,
		}, roundLabels),
		ProcessRound3ComplainCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cylinder_process_round3_complain_count",
			Help: "Number of process round 3 complain",
		}, roundLabels),
		ProcessRound3FailureCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cylinder_process_round3_failure_count",
			Help: "Number of failed process round 3",
		}, roundLabels),
		ProcessRound3Time: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "cylinder_process_round3_time",
			Help: "Time taken to process round 3",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		}, roundLabels),
		DKGLeftGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "cylinder_dkg_left_gauge",
			Help:        "Number of DKG left in the store",
			ConstLabels: labels,
		}),
		GroupCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "cylinder_group_count",
			Help:        "Number of groups in the store",
			ConstLabels: labels,
		}),
		MemberStatusGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "cylinder_member_status",
			Help:        "Status of a member",
			ConstLabels: labels,
		}, memberLabels),
		OnChinDELeftGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "cylinder_on_chain_de_left_gauge",
			Help:        "Number of on-chain DE left",
			ConstLabels: labels,
		}),
		OffChainDELeftGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "cylinder_off_chain_de_left_gauge",
			Help:        "Number of DE left in the store",
			ConstLabels: labels,
		}),
		IncomingSigningCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_incoming_signing_count",
			Help:        "Number of incoming signing requests",
			ConstLabels: labels,
		}, signingLabels),
		ProcessSigningSuccessCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_signing_success_count",
			Help:        "Number of successful process signing",
			ConstLabels: labels,
		}, signingLabels),
		ProcessSigningFailureCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "cylinder_process_signing_failure_count",
			Help:        "Number of failed process signing",
			ConstLabels: labels,
		}, signingLabels),
		ProcessSigningTime: promauto.NewSummaryVec(prometheus.SummaryOpts{
			Name: "cylinder_process_signing_time",
			Help: "Time taken to process signing",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			ConstLabels: labels,
		}, signingLabels),
		WaitingSenderTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "cylinder_waiting_sender_time",
			Help: "Time taken to wait for a free key",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			ConstLabels: labels,
		}),
		SubmittingTxCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "cylinder_submitting_tx_count",
			Help:        "Number of submitting transactions",
			ConstLabels: labels,
		}),
		SubmitTxSuccessCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "cylinder_submit_tx_success_count",
			Help:        "Number of successful submit transactions",
			ConstLabels: labels,
		}),
		SubmitTxFailedCount: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "cylinder_submit_tx_failed_count",
			Help:        "Number of failed submit transactions",
			ConstLabels: labels,
		}),
		SubmitTxTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: "cylinder_submit_tx_time",
			Help: "Time taken to submit transactions",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			ConstLabels: labels,
		}),
	}
}

// StartServer starts a metrics server in a background goroutine, accepting connections
// on the given listener. Any HTTP logging will be written at info level to the given logger.
// The server will be forcefully shut down when ctx finishes.
func StartServer(ctx context.Context, logger *logger.Logger, config *cylinderctx.Config) error {
	ln, err := net.Listen("tcp", config.MetricsListenAddr)
	if err != nil {
		logger.Error(
			"Failed to start metrics server you can change the address and port using metrics-listen-addr config setting or --metrics-listen-addr flag",
		)

		return fmt.Errorf("failed to listen on metrics address %q: %w", config.MetricsListenAddr, err)
	}

	labels := prometheus.Labels{
		"chain_id": config.ChainID,
		"granter":  config.Granter,
	}

	fmt.Printf("labels %+v\n", labels)

	// allow for the global telemetry enabled state to be set.
	globalTelemetryEnabled = true

	// initialize Prometheus metrics
	InitPrometheusMetrics(labels)

	logger.Info("Metrics server listening on address %s", config.MetricsListenAddr)

	// Serve default prometheus metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler:           mux,
		Addr:              config.MetricsListenAddr,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		_ = srv.Serve(ln)
	}()

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	return nil
}
