package grogu

import (
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type groguCollector struct {
	context                   *Context
	reportsHandlingGaugeDesc  *prometheus.Desc
	reportsPendingGaugeDesc   *prometheus.Desc
	reportsErrorCountDesc     *prometheus.Desc
	reportsSubmittedCountDesc *prometheus.Desc
}

func NewGroguCollector(c *Context) prometheus.Collector {
	return &groguCollector{
		context: c,
		reportsHandlingGaugeDesc: prometheus.NewDesc(
			"grogu_reports_handling_count",
			"Number of reports currently being handled",
			nil, nil),
		reportsPendingGaugeDesc: prometheus.NewDesc(
			"grogu_reports_pending_count",
			"Number of reports currently pending for submission",
			nil, nil),
		reportsErrorCountDesc: prometheus.NewDesc(
			"grogu_reports_error_total",
			"Number of report errors since last grogu restart",
			nil, nil),
		reportsSubmittedCountDesc: prometheus.NewDesc(
			"grogu_reports_submitted_total",
			"Number of reports submitted since last grogu restart",
			nil, nil),
	}
}

func (collector groguCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.reportsHandlingGaugeDesc
	ch <- collector.reportsPendingGaugeDesc
	ch <- collector.reportsErrorCountDesc
	ch <- collector.reportsSubmittedCountDesc
}

func (collector groguCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(collector.reportsHandlingGaugeDesc, prometheus.GaugeValue,
		float64(atomic.LoadInt64(&collector.context.handlingGauge)))
	ch <- prometheus.MustNewConstMetric(collector.reportsPendingGaugeDesc, prometheus.GaugeValue,
		float64(atomic.LoadInt64(&collector.context.pendingGauge)))
	ch <- prometheus.MustNewConstMetric(collector.reportsErrorCountDesc, prometheus.CounterValue,
		float64(atomic.LoadInt64(&collector.context.errorCount)))
	ch <- prometheus.MustNewConstMetric(collector.reportsSubmittedCountDesc, prometheus.CounterValue,
		float64(atomic.LoadInt64(&collector.context.submittedCount)))
}

func metricsListen(listenAddr string, c *Context) {
	collector := NewGroguCollector(c)
	prometheus.MustRegister(collector)
	http.Handle("/metrics", promhttp.Handler())
	panic(http.ListenAndServe(listenAddr, nil))
}
