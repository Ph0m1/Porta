package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"

	"github.com/aviddiviner/gin-limit"
	"github.com/gin-gonic/contrib/cache"
	"github.com/gin-gonic/contrib/secure"

	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/config/viper"
	"github.com/ph0m1/porta/logging"
	"github.com/ph0m1/porta/logging/gologging"
	"github.com/ph0m1/porta/proxy"
	pgin "github.com/ph0m1/porta/router/gin"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	debug := flag.Bool("d", true, "Enable the debug")
	configFile := flag.String("c", "configuration.json", "Path to the configuration filename")
	flag.Parse()

	parser := viper.New()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	logger, err := gologging.NewLogger(*logLevel, os.Stdout, "[X_X]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	//routerFactory := gin.DefaultFactory(proxy.DefaultFactory(logger), logger)

	store := cache.NewInMemoryStore(time.Minute)

	mws := []gin.HandlerFunc{
		secure.Secure(secure.Options{
			AllowedHosts:          []string{"127.0.0.1:8080", "example.com", "ssl.example.com"},
			SSLRedirect:           false,
			SSLHost:               "ssl.example.com",
			SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
			STSSeconds:            315360000,
			STSIncludeSubdomains:  true,
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			BrowserXssFilter:      true,
			ContentSecurityPolicy: "default-src 'self'",
		}),
		limit.MaxAllowed(20),
	}

	routerFactory := pgin.NewFactory(pgin.Config{
		Engine:       gin.Default(),
		ProxyFactory: customProxyFactory{logger, proxy.DefaultFactory(logger)},
		Middlewares:  mws,
		Logger:       logger,
		HandlerFactory: func(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
			return cache.CachePage(store, configuration.CacheTTL, pgin.EndpointHandler(configuration, proxy))
		},
	})

	routerFactory.New().Run(serviceConfig)
}

type customProxyFactory struct {
	logger  logging.Logger
	factory proxy.Factory
}

func (cf customProxyFactory) New(cfg *config.EndpointConfig) (p proxy.Proxy, err error) {
	p, err = cf.factory.New(cfg)
	if err == nil {
		p = proxy.NewLoggingMiddleware(cf.logger, cfg.Endpoint)(p)
	}
	return
}
