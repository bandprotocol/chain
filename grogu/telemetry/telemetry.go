package telemetry

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bandprotocol/chain/v3/pkg/logger"
)

// StartServer starts a metrics server in a background goroutine, accepting connections
// on the given listener. Any HTTP logging will be written at info level to the given logger.
// The server will be forcefully shut down when ctx finishes.
func StartServer(l *logger.Logger, metricsListenAddr string) {
	// Initialize the global collector
	collector = NewGroguCollector()
	prometheus.MustRegister(collector)

	// Serve default prometheus metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Handler:           mux,
		Addr:              metricsListenAddr,
		ReadHeaderTimeout: 10 * time.Second,
	}

	l.Info("Metrics server listening on address %s", metricsListenAddr)
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
