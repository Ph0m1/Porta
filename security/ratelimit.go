package security

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
}

// RateLimiter interface defines rate limiting behavior
type RateLimiter interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
	Reset(key string)
	GetStats(key string) RateLimitStats
}

// RateLimitStats holds statistics for a rate limit key
type RateLimitStats struct {
	Requests    int       `json:"requests"`
	Remaining   int       `json:"remaining"`
	ResetTime   time.Time `json:"reset_time"`
	WindowStart time.Time `json:"window_start"`
}

// TokenBucketLimiter implements token bucket rate limiting
type TokenBucketLimiter struct {
	config  *RateLimitConfig
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
	stopCh  chan struct{}
}

type tokenBucket struct {
	tokens      float64
	lastUpdate  time.Time
	requests    int
	windowStart time.Time
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(config *RateLimitConfig) *TokenBucketLimiter {
	limiter := &TokenBucketLimiter{
		config:  config,
		buckets: make(map[string]*tokenBucket),
		stopCh:  make(chan struct{}),
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a single request is allowed
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	return tbl.AllowN(key, 1)
}

// AllowN checks if n requests are allowed
func (tbl *TokenBucketLimiter) AllowN(key string, n int) bool {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	bucket, exists := tbl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:      float64(tbl.config.BurstSize),
			lastUpdate:  time.Now(),
			windowStart: time.Now(),
		}
		tbl.buckets[key] = bucket
	}

	now := time.Now()

	// Reset window if needed
	if now.Sub(bucket.windowStart) >= tbl.config.WindowSize {
		bucket.requests = 0
		bucket.windowStart = now
	}

	// Add tokens based on time elapsed
	elapsed := now.Sub(bucket.lastUpdate)
	tokensToAdd := elapsed.Seconds() * float64(tbl.config.RequestsPerSecond)
	bucket.tokens = min(bucket.tokens+tokensToAdd, float64(tbl.config.BurstSize))
	bucket.lastUpdate = now

	// Check if we have enough tokens
	if bucket.tokens >= float64(n) {
		bucket.tokens -= float64(n)
		bucket.requests += n
		return true
	}

	return false
}

// Reset resets the rate limit for a key
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()
	delete(tbl.buckets, key)
}

// GetStats returns statistics for a key
func (tbl *TokenBucketLimiter) GetStats(key string) RateLimitStats {
	tbl.mu.RLock()
	defer tbl.mu.RUnlock()

	bucket, exists := tbl.buckets[key]
	if !exists {
		return RateLimitStats{
			Remaining: tbl.config.BurstSize,
			ResetTime: time.Now().Add(tbl.config.WindowSize),
		}
	}

	remaining := int(bucket.tokens)
	resetTime := bucket.windowStart.Add(tbl.config.WindowSize)

	return RateLimitStats{
		Requests:    bucket.requests,
		Remaining:   remaining,
		ResetTime:   resetTime,
		WindowStart: bucket.windowStart,
	}
}

// Stop stops the rate limiter
func (tbl *TokenBucketLimiter) Stop() {
	close(tbl.stopCh)
}

// cleanup removes old buckets
func (tbl *TokenBucketLimiter) cleanup() {
	ticker := time.NewTicker(tbl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tbl.mu.Lock()
			now := time.Now()
			for key, bucket := range tbl.buckets {
				if now.Sub(bucket.lastUpdate) > tbl.config.WindowSize*2 {
					delete(tbl.buckets, key)
				}
			}
			tbl.mu.Unlock()
		case <-tbl.stopCh:
			return
		}
	}
}

// SlidingWindowLimiter implements sliding window rate limiting
type SlidingWindowLimiter struct {
	config  *RateLimitConfig
	windows map[string]*slidingWindow
	mu      sync.RWMutex
	stopCh  chan struct{}
}

type slidingWindow struct {
	requests    []time.Time
	totalCount  int
	windowStart time.Time
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(config *RateLimitConfig) *SlidingWindowLimiter {
	limiter := &SlidingWindowLimiter{
		config:  config,
		windows: make(map[string]*slidingWindow),
		stopCh:  make(chan struct{}),
	}

	go limiter.cleanup()
	return limiter
}

// Allow checks if a single request is allowed
func (swl *SlidingWindowLimiter) Allow(key string) bool {
	return swl.AllowN(key, 1)
}

// AllowN checks if n requests are allowed
func (swl *SlidingWindowLimiter) AllowN(key string, n int) bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	window, exists := swl.windows[key]
	if !exists {
		window = &slidingWindow{
			requests:    make([]time.Time, 0),
			windowStart: time.Now(),
		}
		swl.windows[key] = window
	}

	now := time.Now()
	windowStart := now.Add(-swl.config.WindowSize)

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range window.requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}
	window.requests = validRequests

	// Check if we can add n more requests
	if len(window.requests)+n <= swl.config.RequestsPerSecond {
		for i := 0; i < n; i++ {
			window.requests = append(window.requests, now)
		}
		window.totalCount += n
		return true
	}

	return false
}

// Reset resets the rate limit for a key
func (swl *SlidingWindowLimiter) Reset(key string) {
	swl.mu.Lock()
	defer swl.mu.Unlock()
	delete(swl.windows, key)
}

// GetStats returns statistics for a key
func (swl *SlidingWindowLimiter) GetStats(key string) RateLimitStats {
	swl.mu.RLock()
	defer swl.mu.RUnlock()

	window, exists := swl.windows[key]
	if !exists {
		return RateLimitStats{
			Remaining: swl.config.RequestsPerSecond,
			ResetTime: time.Now().Add(swl.config.WindowSize),
		}
	}

	now := time.Now()
	windowStart := now.Add(-swl.config.WindowSize)

	// Count valid requests
	validCount := 0
	for _, reqTime := range window.requests {
		if reqTime.After(windowStart) {
			validCount++
		}
	}

	remaining := swl.config.RequestsPerSecond - validCount
	if remaining < 0 {
		remaining = 0
	}

	return RateLimitStats{
		Requests:    validCount,
		Remaining:   remaining,
		ResetTime:   windowStart.Add(swl.config.WindowSize),
		WindowStart: windowStart,
	}
}

// Stop stops the rate limiter
func (swl *SlidingWindowLimiter) Stop() {
	close(swl.stopCh)
}

// cleanup removes old windows
func (swl *SlidingWindowLimiter) cleanup() {
	ticker := time.NewTicker(swl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			swl.mu.Lock()
			now := time.Now()
			for key, window := range swl.windows {
				if now.Sub(window.windowStart) > swl.config.WindowSize*2 {
					delete(swl.windows, key)
				}
			}
			swl.mu.Unlock()
		case <-swl.stopCh:
			return
		}
	}
}

// RateLimitMiddleware provides rate limiting middleware
type RateLimitMiddleware struct {
	limiter RateLimiter
	keyFunc func(*http.Request) string
	onLimit func(http.ResponseWriter, *http.Request, RateLimitStats)
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(limiter RateLimiter, keyFunc func(*http.Request) string) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
		keyFunc: keyFunc,
		onLimit: defaultOnLimit,
	}
}

// SetOnLimit sets the function to call when rate limit is exceeded
func (rlm *RateLimitMiddleware) SetOnLimit(onLimit func(http.ResponseWriter, *http.Request, RateLimitStats)) {
	rlm.onLimit = onLimit
}

// HTTPMiddleware returns an HTTP middleware function
func (rlm *RateLimitMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rlm.keyFunc(r)

		if !rlm.limiter.Allow(key) {
			stats := rlm.limiter.GetStats(key)
			rlm.onLimit(w, r, stats)
			return
		}

		// Add rate limit headers
		stats := rlm.limiter.GetStats(key)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(stats.Requests+stats.Remaining))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(stats.Remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(stats.ResetTime.Unix(), 10))

		next.ServeHTTP(w, r)
	})
}

// defaultOnLimit is the default handler for rate limit exceeded
func defaultOnLimit(w http.ResponseWriter, r *http.Request, stats RateLimitStats) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(stats.Requests+stats.Remaining))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(stats.ResetTime.Unix(), 10))
	w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(stats.ResetTime).Seconds()), 10))

	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, `{"error":"rate limit exceeded","retry_after":%d}`, int64(time.Until(stats.ResetTime).Seconds()))
}

// Common key functions

// IPKeyFunc creates a key function based on client IP
func IPKeyFunc(r *http.Request) string {
	// Try to get real IP from headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return "ip:" + ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return "ip:" + ip
	}
	return "ip:" + r.RemoteAddr
}

// UserKeyFunc creates a key function based on authenticated user
func UserKeyFunc(r *http.Request) string {
	if authCtx, ok := GetAuthContext(r); ok {
		if authCtx.UserID != "" {
			return "user:" + authCtx.UserID
		}
		if authCtx.ClientID != "" {
			return "client:" + authCtx.ClientID
		}
	}
	return IPKeyFunc(r)
}

// EndpointKeyFunc creates a key function based on endpoint and user
func EndpointKeyFunc(r *http.Request) string {
	endpoint := r.URL.Path
	userKey := UserKeyFunc(r)
	return fmt.Sprintf("%s:%s", userKey, endpoint)
}

// Helper function for min
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
