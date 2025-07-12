package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/ph0m1/porta/config"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string                                 `json:"name"`
	Status      HealthStatus                           `json:"status"`
	Message     string                                 `json:"message,omitempty"`
	LastChecked time.Time                              `json:"last_checked"`
	Duration    time.Duration                          `json:"duration"`
	CheckFunc   func(ctx context.Context) HealthResult `json:"-"`
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Status  HealthStatus
	Message string
}

// HealthChecker manages all health checks
type HealthChecker struct {
	checks   map[string]*HealthCheck
	mu       sync.RWMutex
	interval time.Duration
	timeout  time.Duration
	stopCh   chan struct{}
}

// OverallHealth represents the overall health status
type OverallHealth struct {
	Status     HealthStatus           `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version,omitempty"`
	Uptime     time.Duration          `json:"uptime"`
	Checks     map[string]HealthCheck `json:"checks"`
	SystemInfo SystemInfo             `json:"system_info"`
}

// SystemInfo contains system-level information
type SystemInfo struct {
	Goroutines  int    `json:"goroutines"`
	MemoryAlloc uint64 `json:"memory_alloc_bytes"`
	MemorySys   uint64 `json:"memory_sys_bytes"`
	CPUCount    int    `json:"cpu_count"`
	GoVersion   string `json:"go_version"`
}

var startTime = time.Now()

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		checks:   make(map[string]*HealthCheck),
		interval: interval,
		timeout:  timeout,
		stopCh:   make(chan struct{}),
	}
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, checkFunc func(ctx context.Context) HealthResult) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = &HealthCheck{
		Name:      name,
		Status:    StatusHealthy,
		CheckFunc: checkFunc,
	}
}

// Start begins the health checking routine
func (hc *HealthChecker) Start() {
	go hc.runChecks()
}

// Stop stops the health checking routine
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// GetHealth returns the current health status
func (hc *HealthChecker) GetHealth() OverallHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	checks := make(map[string]HealthCheck)
	overallStatus := StatusHealthy

	for name, check := range hc.checks {
		checks[name] = *check
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return OverallHealth{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
		Checks:    checks,
		SystemInfo: SystemInfo{
			Goroutines:  runtime.NumGoroutine(),
			MemoryAlloc: m.Alloc,
			MemorySys:   m.Sys,
			CPUCount:    runtime.NumCPU(),
			GoVersion:   runtime.Version(),
		},
	}
}

// runChecks runs all health checks periodically
func (hc *HealthChecker) runChecks() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// Run initial checks
	hc.executeChecks()

	for {
		select {
		case <-ticker.C:
			hc.executeChecks()
		case <-hc.stopCh:
			return
		}
	}
}

// executeChecks executes all registered health checks
func (hc *HealthChecker) executeChecks() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, check := range hc.checks {
		go hc.executeCheck(check)
	}
}

// executeCheck executes a single health check
func (hc *HealthChecker) executeCheck(check *HealthCheck) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	result := check.CheckFunc(ctx)

	check.Status = result.Status
	check.Message = result.Message
	check.LastChecked = time.Now()
	check.Duration = time.Since(start)
}

// HTTPHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := hc.GetHealth()

		w.Header().Set("Content-Type", "application/json")

		// Set HTTP status based on health
		switch health.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK, but degraded
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(health)
	}
}

// ReadinessHandler returns a simple readiness check handler
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := hc.GetHealth()

		if health.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Not Ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	}
}

// LivenessHandler returns a simple liveness check handler
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Alive"))
	}
}

// CreateDefaultHealthChecks creates default health checks for the gateway
func CreateDefaultHealthChecks(serviceConfig *config.ServiceConfig) *HealthChecker {
	hc := NewHealthChecker(30*time.Second, 5*time.Second)

	// Memory usage check
	hc.RegisterCheck("memory", func(ctx context.Context) HealthResult {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Alert if memory usage is over 1GB
		if m.Alloc > 1024*1024*1024 {
			return HealthResult{
				Status:  StatusDegraded,
				Message: fmt.Sprintf("High memory usage: %d MB", m.Alloc/1024/1024),
			}
		}

		return HealthResult{
			Status:  StatusHealthy,
			Message: fmt.Sprintf("Memory usage: %d MB", m.Alloc/1024/1024),
		}
	})

	// Goroutine count check
	hc.RegisterCheck("goroutines", func(ctx context.Context) HealthResult {
		count := runtime.NumGoroutine()

		// Alert if goroutine count is over 1000
		if count > 1000 {
			return HealthResult{
				Status:  StatusDegraded,
				Message: fmt.Sprintf("High goroutine count: %d", count),
			}
		}

		return HealthResult{
			Status:  StatusHealthy,
			Message: fmt.Sprintf("Goroutine count: %d", count),
		}
	})

	// Backend connectivity check
	for _, endpoint := range serviceConfig.Endpoints {
		for i, backend := range endpoint.Backend {
			backendName := fmt.Sprintf("backend_%s_%d", endpoint.Endpoint, i)
			hosts := backend.Host

			hc.RegisterCheck(backendName, func(ctx context.Context) HealthResult {
				client := &http.Client{Timeout: 3 * time.Second}

				for _, host := range hosts {
					req, err := http.NewRequestWithContext(ctx, "GET", host+"/__health", nil)
					if err != nil {
						continue
					}

					resp, err := client.Do(req)
					if err != nil {
						continue
					}
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						return HealthResult{
							Status:  StatusHealthy,
							Message: fmt.Sprintf("Backend %s is healthy", host),
						}
					}
				}

				return HealthResult{
					Status:  StatusUnhealthy,
					Message: fmt.Sprintf("All backend hosts are unreachable"),
				}
			})
		}
	}

	return hc
}
