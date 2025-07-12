package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/config/viper"
	"github.com/ph0m1/porta/logging"
	"github.com/ph0m1/porta/logging/gologging"
	"github.com/ph0m1/porta/monitoring"
	"github.com/ph0m1/porta/proxy"
	"github.com/ph0m1/porta/router/gin"
	"github.com/ph0m1/porta/security"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "INFO", "Logging level")
	debug := flag.Bool("d", false, "Enable debug mode")
	configFile := flag.String("c", "../etc/config.yaml", "Path to the configuration filename")
	securityFile := flag.String("s", "../etc/security.yaml", "Path to the security configuration filename")
	flag.Parse()

	// Parse main configuration
	parser := viper.New()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	// Parse security configuration
	securityConfig, err := parseSecurityConfig(*securityFile)
	if err != nil {
		log.Printf("WARNING: Could not load security config: %v", err)
		securityConfig = getDefaultSecurityConfig()
	}

	// Create logger
	logger, err := gologging.NewLogger(*logLevel, os.Stdout, "[PORTA-SECURE]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	// Initialize metrics
	metrics := monitoring.NewMetrics()

	// Initialize health checker
	healthChecker := monitoring.CreateDefaultHealthChecks(&serviceConfig)
	healthChecker.Start()
	defer healthChecker.Stop()

	// Create Gin engine
	if !serviceConfig.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// Add middleware stack
	setupMiddleware(engine, securityConfig, metrics, logger, healthChecker)

	// Create proxy factory with monitoring
	proxyFactory := newMonitoredProxyFactory(proxy.DefaultFactory(logger), metrics, logger)

	// Create router factory
	routerFactory := pgin.NewFactory(pgin.Config{
		Engine:       engine,
		ProxyFactory: proxyFactory,
		Logger:       logger,
		HandlerFactory: func(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
			return newMonitoredHandler(configuration, proxy, metrics)
		},
	})

	// Start the gateway
	logger.Info("Starting Porta Gateway with enhanced security and monitoring...")
	routerFactory.New().Run(serviceConfig)
}

// setupMiddleware configures all middleware
func setupMiddleware(engine *gin.Engine, securityConfig *SecurityConfig, metrics *monitoring.Metrics, logger logging.Logger, healthChecker *monitoring.HealthChecker) {
	// Recovery middleware
	engine.Use(gin.Recovery())

	// Request ID middleware
	requestIDMiddleware := security.NewRequestIDMiddleware("X-Request-ID")
	engine.Use(gin.WrapH(requestIDMiddleware.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))

	// Security headers middleware
	securityHeadersMiddleware := security.NewSecurityHeadersMiddleware(&security.SecurityHeadersConfig{
		ContentTypeNosniff:    securityConfig.SecurityHeaders.ContentTypeNosniff,
		FrameDeny:             securityConfig.SecurityHeaders.FrameDeny,
		BrowserXSSFilter:      securityConfig.SecurityHeaders.BrowserXSSFilter,
		ContentSecurityPolicy: securityConfig.SecurityHeaders.ContentSecurityPolicy,
		ReferrerPolicy:        securityConfig.SecurityHeaders.ReferrerPolicy,
		HSTSMaxAge:            securityConfig.SecurityHeaders.HSTSMaxAge,
		HSTSIncludeSubdomains: securityConfig.SecurityHeaders.HSTSIncludeSubdomains,
		HSTSPreload:           securityConfig.SecurityHeaders.HSTSPreload,
	})
	engine.Use(gin.WrapH(securityHeadersMiddleware.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))

	// CORS middleware
	corsMiddleware := security.NewCORSMiddleware(&security.CORSConfig{
		AllowedOrigins:   securityConfig.CORS.AllowedOrigins,
		AllowedMethods:   securityConfig.CORS.AllowedMethods,
		AllowedHeaders:   securityConfig.CORS.AllowedHeaders,
		ExposedHeaders:   securityConfig.CORS.ExposedHeaders,
		AllowCredentials: securityConfig.CORS.AllowCredentials,
		MaxAge:           securityConfig.CORS.MaxAge,
	})
	engine.Use(gin.WrapH(corsMiddleware.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))

	// Rate limiting middleware
	rateLimiter := security.NewTokenBucketLimiter(&security.RateLimitConfig{
		RequestsPerSecond: securityConfig.RateLimit.RequestsPerSecond,
		BurstSize:         securityConfig.RateLimit.BurstSize,
		WindowSize:        time.Duration(securityConfig.RateLimit.WindowSize) * time.Second,
		CleanupInterval:   time.Duration(securityConfig.RateLimit.CleanupInterval) * time.Second,
	})
	rateLimitMiddleware := security.NewRateLimitMiddleware(rateLimiter, security.UserKeyFunc)
	engine.Use(gin.WrapH(rateLimitMiddleware.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))

	// Authentication middleware (optional)
	if securityConfig.Auth.Enabled {
		authMiddleware := security.NewAuthMiddleware(&security.AuthConfig{
			JWTSecret:     securityConfig.Auth.JWTSecret,
			JWTExpiration: time.Duration(securityConfig.Auth.JWTExpiration) * time.Hour,
			APIKeys:       securityConfig.Auth.APIKeys,
			BasicAuth:     securityConfig.Auth.BasicAuth,
			RequiredRoles: securityConfig.Auth.RequiredRoles,
		})
		engine.Use(gin.WrapH(authMiddleware.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
	}

	// Request logging middleware
	engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return logger.Info(
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)
		return ""
	}))

	// Add monitoring endpoints
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.GET("/__health", gin.WrapH(healthChecker.HTTPHandler()))
	engine.GET("/__ready", gin.WrapH(healthChecker.ReadinessHandler()))
	engine.GET("/__live", gin.WrapH(monitoring.LivenessHandler()))

	// Add admin endpoints (if auth is enabled)
	if securityConfig.Auth.Enabled {
		adminGroup := engine.Group("/admin")
		adminGroup.GET("/metrics", gin.WrapH(promhttp.Handler()))
		adminGroup.GET("/health", gin.WrapH(healthChecker.HTTPHandler()))
	}
}

// newMonitoredProxyFactory creates a proxy factory with monitoring
func newMonitoredProxyFactory(factory proxy.Factory, metrics *monitoring.Metrics, logger logging.Logger) proxy.Factory {
	return &monitoredProxyFactory{
		factory: factory,
		metrics: metrics,
		logger:  logger,
	}
}

type monitoredProxyFactory struct {
	factory proxy.Factory
	metrics *monitoring.Metrics
	logger  logging.Logger
}

func (mpf *monitoredProxyFactory) New(cfg *config.EndpointConfig) (proxy.Proxy, error) {
	p, err := mpf.factory.New(cfg)
	if err != nil {
		return nil, err
	}

	// Wrap proxy with monitoring
	return proxy.NewLoggingMiddleware(mpf.logger, cfg.Endpoint)(p), nil
}

// newMonitoredHandler creates a handler with monitoring
func newMonitoredHandler(cfg *config.EndpointConfig, p proxy.Proxy, metrics *monitoring.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		metrics.IncRequestsInFlight(c.Request.Method, cfg.Endpoint)
		defer metrics.DecRequestsInFlight(c.Request.Method, cfg.Endpoint)

		// Call the original handler
		pgin.EndpointHandler(cfg, p)(c)

		// Record metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		requestSize := c.Request.ContentLength
		responseSize := int64(c.Writer.Size())

		metrics.RecordRequest(
			c.Request.Method,
			cfg.Endpoint,
			string(rune(statusCode)),
			duration,
			requestSize,
			responseSize,
		)
	}
}

// SecurityConfig represents the security configuration structure
type SecurityConfig struct {
	Auth struct {
		Enabled       bool                `yaml:"enabled"`
		JWTSecret     string              `yaml:"jwt_secret"`
		JWTExpiration int                 `yaml:"jwt_expiration"`
		APIKeys       map[string]string   `yaml:"api_keys"`
		BasicAuth     map[string]string   `yaml:"basic_auth"`
		RequiredRoles map[string][]string `yaml:"required_roles"`
	} `yaml:"auth"`

	RateLimit struct {
		RequestsPerSecond int `yaml:"requests_per_second"`
		BurstSize         int `yaml:"burst_size"`
		WindowSize        int `yaml:"window_size"`
		CleanupInterval   int `yaml:"cleanup_interval"`
	} `yaml:"rate_limit"`

	CORS struct {
		AllowedOrigins   []string `yaml:"allowed_origins"`
		AllowedMethods   []string `yaml:"allowed_methods"`
		AllowedHeaders   []string `yaml:"allowed_headers"`
		ExposedHeaders   []string `yaml:"exposed_headers"`
		AllowCredentials bool     `yaml:"allow_credentials"`
		MaxAge           int      `yaml:"max_age"`
	} `yaml:"cors"`

	SecurityHeaders struct {
		ContentTypeNosniff    bool   `yaml:"content_type_nosniff"`
		FrameDeny             bool   `yaml:"frame_deny"`
		BrowserXSSFilter      bool   `yaml:"browser_xss_filter"`
		ContentSecurityPolicy string `yaml:"content_security_policy"`
		ReferrerPolicy        string `yaml:"referrer_policy"`
		HSTSMaxAge            int    `yaml:"hsts_max_age"`
		HSTSIncludeSubdomains bool   `yaml:"hsts_include_subdomains"`
		HSTSPreload           bool   `yaml:"hsts_preload"`
	} `yaml:"security_headers"`
}

// parseSecurityConfig parses the security configuration file
func parseSecurityConfig(filename string) (*SecurityConfig, error) {
	parser := viper.New()
	// This is a simplified version - in reality you'd need to implement
	// proper YAML parsing for the security config
	return getDefaultSecurityConfig(), nil
}

// getDefaultSecurityConfig returns a default security configuration
func getDefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Auth: struct {
			Enabled       bool                `yaml:"enabled"`
			JWTSecret     string              `yaml:"jwt_secret"`
			JWTExpiration int                 `yaml:"jwt_expiration"`
			APIKeys       map[string]string   `yaml:"api_keys"`
			BasicAuth     map[string]string   `yaml:"basic_auth"`
			RequiredRoles map[string][]string `yaml:"required_roles"`
		}{
			Enabled:       false,
			JWTSecret:     "default-secret-change-in-production",
			JWTExpiration: 24,
			APIKeys:       make(map[string]string),
			BasicAuth:     make(map[string]string),
			RequiredRoles: make(map[string][]string),
		},
		RateLimit: struct {
			RequestsPerSecond int `yaml:"requests_per_second"`
			BurstSize         int `yaml:"burst_size"`
			WindowSize        int `yaml:"window_size"`
			CleanupInterval   int `yaml:"cleanup_interval"`
		}{
			RequestsPerSecond: 100,
			BurstSize:         200,
			WindowSize:        60,
			CleanupInterval:   300,
		},
		CORS: struct {
			AllowedOrigins   []string `yaml:"allowed_origins"`
			AllowedMethods   []string `yaml:"allowed_methods"`
			AllowedHeaders   []string `yaml:"allowed_headers"`
			ExposedHeaders   []string `yaml:"exposed_headers"`
			AllowCredentials bool     `yaml:"allow_credentials"`
			MaxAge           int      `yaml:"max_age"`
		}{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
			AllowCredentials: false,
			MaxAge:           86400,
		},
		SecurityHeaders: struct {
			ContentTypeNosniff    bool   `yaml:"content_type_nosniff"`
			FrameDeny             bool   `yaml:"frame_deny"`
			BrowserXSSFilter      bool   `yaml:"browser_xss_filter"`
			ContentSecurityPolicy string `yaml:"content_security_policy"`
			ReferrerPolicy        string `yaml:"referrer_policy"`
			HSTSMaxAge            int    `yaml:"hsts_max_age"`
			HSTSIncludeSubdomains bool   `yaml:"hsts_include_subdomains"`
			HSTSPreload           bool   `yaml:"hsts_preload"`
		}{
			ContentTypeNosniff:    true,
			FrameDeny:             true,
			BrowserXSSFilter:      true,
			ContentSecurityPolicy: "default-src 'self'",
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			HSTSMaxAge:            31536000,
			HSTSIncludeSubdomains: true,
			HSTSPreload:           false,
		},
	}
}
