package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"multi-protocol-proxy/internal/ui"
)

var (
	MetricRelayTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "multiproxy_relay_total",
		Help: "Total relayed connections by SNI",
	}, []string{"sni"})

	MetricActiveConns = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "multiproxy_active_conns",
		Help: "Current active connections",
	})

	MetricBytesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "multiproxy_bytes_total",
		Help: "Total bytes transferred",
	}, []string{"sni", "direction"})

	MetricErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "multiproxy_errors_total",
		Help: "Total errors by type",
	}, []string{"type"})

	MetricConnectionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "multiproxy_connection_duration_seconds",
		Help:    "Connection duration in seconds",
		Buckets: []float64{1, 5, 15, 30, 60, 120, 300, 600},
	})

	MetricConnectionsRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "multiproxy_connections_rejected_total",
		Help: "Total connections rejected due to capacity",
	})
)

var activeConnsMu sync.Mutex
var activeConnsCount int

func init() {
	origInc := MetricActiveConns.Inc
	origDec := MetricActiveConns.Dec
	MetricActiveConns = &gaugeWrapper{
		Gauge:  MetricActiveConns.(prometheus.Gauge),
		inc:    origInc,
		dec:    origDec,
		count:  &activeConnsCount,
		mu:     &activeConnsMu,
	}
}

type gaugeWrapper struct {
	prometheus.Gauge
	inc   func()
	dec   func()
	count *int
	mu    *sync.Mutex
}

func (g *gaugeWrapper) Inc() {
	g.mu.Lock()
	*g.count++
	g.mu.Unlock()
	g.Gauge.Inc()
}

func (g *gaugeWrapper) Dec() {
	g.mu.Lock()
	*g.count--
	g.mu.Unlock()
	g.Gauge.Dec()
}

func (g *gaugeWrapper) Get() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return *g.count
}

func GetActiveConns() int {
	activeConnsMu.Lock()
	defer activeConnsMu.Unlock()
	return activeConnsCount
}

type MetricsServer struct {
	server *http.Server
}

func NewMetricsServer(addr string, usageHandler http.HandlerFunc) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api/stats", StatsHandler)
	mux.HandleFunc("/api/history", HistoryHandler)
	if usageHandler != nil {
		mux.HandleFunc("/api/usage", usageHandler)
	}

	return &MetricsServer{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (m *MetricsServer) Start() {
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ui.LogStatus("error", "Metrics server error: "+err.Error())
		}
	}()
}

func (m *MetricsServer) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.server.Shutdown(shutdownCtx)
}
