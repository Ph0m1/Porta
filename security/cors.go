package security

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins     []string `json:"allowed_origins"`
	AllowedMethods     []string `json:"allowed_methods"`
	AllowedHeaders     []string `json:"allowed_headers"`
	ExposedHeaders     []string `json:"exposed_headers"`
	AllowCredentials   bool     `json:"allow_credentials"`
	MaxAge             int      `json:"max_age"`
	OptionsPassthrough bool     `json:"options_passthrough"`
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Language",
			"Content-Type",
			"Authorization",
			"X-API-Key",
			"X-Client-ID",
			"X-Signature",
			"X-Timestamp",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware provides CORS middleware
type CORSMiddleware struct {
	config *CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig()
	}
	return &CORSMiddleware{config: config}
}

// HTTPMiddleware returns an HTTP middleware function
func (cm *CORSMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if cm.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(cm.config.AllowedOrigins) == 1 && cm.config.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		if len(cm.config.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.config.AllowedMethods, ", "))
		}

		if len(cm.config.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.config.AllowedHeaders, ", "))
		}

		if len(cm.config.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cm.config.ExposedHeaders, ", "))
		}

		if cm.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if cm.config.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cm.config.MaxAge))
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			if !cm.config.OptionsPassthrough {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the origin is allowed
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	ContentTypeNosniff    bool   `json:"content_type_nosniff"`
	FrameDeny             bool   `json:"frame_deny"`
	FrameOptions          string `json:"frame_options"`
	BrowserXSSFilter      bool   `json:"browser_xss_filter"`
	ContentSecurityPolicy string `json:"content_security_policy"`
	ReferrerPolicy        string `json:"referrer_policy"`
	FeaturePolicy         string `json:"feature_policy"`
	PermissionsPolicy     string `json:"permissions_policy"`
	HSTSMaxAge            int    `json:"hsts_max_age"`
	HSTSIncludeSubdomains bool   `json:"hsts_include_subdomains"`
	HSTSPreload           bool   `json:"hsts_preload"`
}

// DefaultSecurityHeadersConfig returns a default security headers configuration
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		BrowserXSSFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	}
}

// SecurityHeadersMiddleware provides security headers middleware
type SecurityHeadersMiddleware struct {
	config *SecurityHeadersConfig
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(config *SecurityHeadersConfig) *SecurityHeadersMiddleware {
	if config == nil {
		config = DefaultSecurityHeadersConfig()
	}
	return &SecurityHeadersMiddleware{config: config}
}

// HTTPMiddleware returns an HTTP middleware function
func (shm *SecurityHeadersMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options
		if shm.config.ContentTypeNosniff {
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if shm.config.FrameDeny {
			w.Header().Set("X-Frame-Options", "DENY")
		} else if shm.config.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", shm.config.FrameOptions)
		}

		// X-XSS-Protection
		if shm.config.BrowserXSSFilter {
			w.Header().Set("X-XSS-Protection", "1; mode=block")
		}

		// Content-Security-Policy
		if shm.config.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", shm.config.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if shm.config.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", shm.config.ReferrerPolicy)
		}

		// Feature-Policy (deprecated, use Permissions-Policy)
		if shm.config.FeaturePolicy != "" {
			w.Header().Set("Feature-Policy", shm.config.FeaturePolicy)
		}

		// Permissions-Policy
		if shm.config.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", shm.config.PermissionsPolicy)
		}

		// HSTS (only for HTTPS)
		if r.TLS != nil && shm.config.HSTSMaxAge > 0 {
			hstsValue := "max-age=" + strconv.Itoa(shm.config.HSTSMaxAge)
			if shm.config.HSTSIncludeSubdomains {
				hstsValue += "; includeSubDomains"
			}
			if shm.config.HSTSPreload {
				hstsValue += "; preload"
			}
			w.Header().Set("Strict-Transport-Security", hstsValue)
		}

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
type RequestIDMiddleware struct {
	header string
}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware(header string) *RequestIDMiddleware {
	if header == "" {
		header = "X-Request-ID"
	}
	return &RequestIDMiddleware{header: header}
}

// HTTPMiddleware returns an HTTP middleware function
func (rim *RequestIDMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(rim.header)
		if requestID == "" {
			requestID = generateRequestID()
		}

		w.Header().Set(rim.header, requestID)
		r.Header.Set(rim.header, requestID)

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds request timeout
type TimeoutMiddleware struct {
	timeout time.Duration
}

// NewTimeoutMiddleware creates a new timeout middleware
func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{timeout: timeout}
}

// HTTPMiddleware returns an HTTP middleware function
func (tm *TimeoutMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, tm.timeout, "Request timeout")
}

// CompressionMiddleware provides response compression
type CompressionMiddleware struct {
	level int
}

// NewCompressionMiddleware creates a new compression middleware
func NewCompressionMiddleware(level int) *CompressionMiddleware {
	return &CompressionMiddleware{level: level}
}

// HTTPMiddleware returns an HTTP middleware function
func (cm *CompressionMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for certain content types
		contentType := w.Header().Get("Content-Type")
		if shouldSkipCompression(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Note: In a real implementation, you would use a proper gzip writer
		// This is a simplified version
		next.ServeHTTP(w, r)
	})
}

// shouldSkipCompression checks if compression should be skipped for the content type
func shouldSkipCompression(contentType string) bool {
	skipTypes := []string{
		"image/",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/x-gzip",
	}

	for _, skipType := range skipTypes {
		if strings.HasPrefix(contentType, skipType) {
			return true
		}
	}
	return false
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// In a real implementation, you would use a proper UUID library
	// This is a simplified version using timestamp
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// IPWhitelistMiddleware provides IP whitelisting
type IPWhitelistMiddleware struct {
	allowedIPs []string
}

// NewIPWhitelistMiddleware creates a new IP whitelist middleware
func NewIPWhitelistMiddleware(allowedIPs []string) *IPWhitelistMiddleware {
	return &IPWhitelistMiddleware{allowedIPs: allowedIPs}
}

// HTTPMiddleware returns an HTTP middleware function
func (iwm *IPWhitelistMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !iwm.isIPAllowed(clientIP) {
			http.Error(w, "Forbidden: IP not allowed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isIPAllowed checks if the IP is in the whitelist
func (iwm *IPWhitelistMiddleware) isIPAllowed(ip string) bool {
	for _, allowedIP := range iwm.allowedIPs {
		if allowedIP == ip || allowedIP == "*" {
			return true
		}
		// Support CIDR notation in a real implementation
	}
	return false
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Try to get real IP from headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(ip, ","); idx != -1 {
			return strings.TrimSpace(ip[:idx])
		}
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return ip
	}

	// Extract IP from RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}
