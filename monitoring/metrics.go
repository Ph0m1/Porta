package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the Prometheus metrics for the gateway
type Metrics struct {
	// Request metrics
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight *prometheus.GaugeVec
	RequestSize      *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec

	// Backend metrics
	BackendRequestsTotal    *prometheus.CounterVec
	BackendRequestDuration  *prometheus.HistogramVec
	BackendRequestsInFlight *prometheus.GaugeVec
	BackendErrors           *prometheus.CounterVec

	// System metrics
	GoroutinesCount prometheus.Gauge
	MemoryUsage     *prometheus.GaugeVec
	CPUUsage        prometheus.Gauge

	// Circuit breaker metrics
	CircuitBreakerState *prometheus.GaugeVec
	CircuitBreakerTrips *prometheus.CounterVec

	// Rate limiting metrics
	RateLimitHits   *prometheus.CounterVec
	RateLimitBlocks *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// Request metrics
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "porta_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status_code"},
		),

		RequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "porta_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"method", "endpoint"},
		),

		RequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "porta_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),

		ResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "porta_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint", "status_code"},
		),

		// Backend metrics
		BackendRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_backend_requests_total",
				Help: "Total number of requests sent to backends",
			},
			[]string{"backend", "method", "status_code"},
		),

		BackendRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "porta_backend_request_duration_seconds",
				Help:    "Backend request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"backend", "method", "status_code"},
		),

		BackendRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "porta_backend_requests_in_flight",
				Help: "Number of requests currently being sent to backends",
			},
			[]string{"backend"},
		),

		BackendErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_backend_errors_total",
				Help: "Total number of backend errors",
			},
			[]string{"backend", "error_type"},
		),

		// System metrics
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "porta_goroutines_count",
				Help: "Number of goroutines currently running",
			},
		),

		MemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "porta_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),

		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "porta_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),

		// Circuit breaker metrics
		CircuitBreakerState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "porta_circuit_breaker_state",
				Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
			},
			[]string{"backend"},
		),

		CircuitBreakerTrips: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_circuit_breaker_trips_total",
				Help: "Total number of circuit breaker trips",
			},
			[]string{"backend"},
		),

		// Rate limiting metrics
		RateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"client_id", "endpoint"},
		),

		RateLimitBlocks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "porta_rate_limit_blocks_total",
				Help: "Total number of rate limit blocks",
			},
			[]string{"client_id", "endpoint"},
		),
	}
}

// RecordRequest records metrics for an HTTP request
func (m *Metrics) RecordRequest(method, endpoint, statusCode string, duration time.Duration, requestSize, responseSize int64) {
	m.RequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.RequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration.Seconds())
	m.RequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	m.ResponseSize.WithLabelValues(method, endpoint, statusCode).Observe(float64(responseSize))
}

// RecordBackendRequest records metrics for a backend request
func (m *Metrics) RecordBackendRequest(backend, method, statusCode string, duration time.Duration) {
	m.BackendRequestsTotal.WithLabelValues(backend, method, statusCode).Inc()
	m.BackendRequestDuration.WithLabelValues(backend, method, statusCode).Observe(duration.Seconds())
}

// RecordBackendError records a backend error
func (m *Metrics) RecordBackendError(backend, errorType string) {
	m.BackendErrors.WithLabelValues(backend, errorType).Inc()
}

// IncRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncRequestsInFlight(method, endpoint string) {
	m.RequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// DecRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecRequestsInFlight(method, endpoint string) {
	m.RequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// IncBackendRequestsInFlight increments the backend in-flight requests counter
func (m *Metrics) IncBackendRequestsInFlight(backend string) {
	m.BackendRequestsInFlight.WithLabelValues(backend).Inc()
}

// DecBackendRequestsInFlight decrements the backend in-flight requests counter
func (m *Metrics) DecBackendRequestsInFlight(backend string) {
	m.BackendRequestsInFlight.WithLabelValues(backend).Dec()
}

// SetCircuitBreakerState sets the circuit breaker state
func (m *Metrics) SetCircuitBreakerState(backend string, state int) {
	m.CircuitBreakerState.WithLabelValues(backend).Set(float64(state))
}

// RecordCircuitBreakerTrip records a circuit breaker trip
func (m *Metrics) RecordCircuitBreakerTrip(backend string) {
	m.CircuitBreakerTrips.WithLabelValues(backend).Inc()
}

// RecordRateLimit records rate limiting metrics
func (m *Metrics) RecordRateLimit(clientID, endpoint string, blocked bool) {
	m.RateLimitHits.WithLabelValues(clientID, endpoint).Inc()
	if blocked {
		m.RateLimitBlocks.WithLabelValues(clientID, endpoint).Inc()
	}
}

// UpdateSystemMetrics updates system-level metrics
func (m *Metrics) UpdateSystemMetrics(goroutines int, memAlloc, memSys uint64, cpuPercent float64) {
	m.GoroutinesCount.Set(float64(goroutines))
	m.MemoryUsage.WithLabelValues("alloc").Set(float64(memAlloc))
	m.MemoryUsage.WithLabelValues("sys").Set(float64(memSys))
	m.CPUUsage.Set(cpuPercent)
}
