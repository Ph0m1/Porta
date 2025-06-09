package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	gorilla "github.com/gorilla/mux"
	"gopkg.in/unrolled/secure.v1"

	"github.com/ph0m1/p_gateway/config/viper"
	"github.com/ph0m1/p_gateway/logging/gologging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router/mux"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Enable the debug")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "configuration.json", "Path to configuration")
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

	logger, err := gologging.NewLogger(*logLevel, os.Stdout, "[PORTA]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())

		secureMiddleware := secure.New(secure.Options{
			AllowedHosts:          []string{"127.0.0.1:8080", "example.com", "ssl.example.com"},
			STSSeconds:            315360000,
			STSIncludeSubdomains:  true,
			STSPreload:            true,
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			BrowserXssFilter:      true,
			ContentSecurityPolicy: "default-src 'self'",
		})

		routerFactory := mux.NewFactory(mux.Config{
			Engine:       gorillaEngine{gorilla.NewRouter()},
			ProxyFactory: proxy.DefaultFactory(logger),
			Middlewares:    []mux.HandlerMiddleware{secureMiddleware},Add commentMore actions
			Logger:         logger,
			HandlerFactory: mux.EndpointHandler,
			DebugPattern:   "/__debug/{params}",
		})

		routerFactory.New().Run(serviceConfig)
	}

	type gorillaEngine struct {
		r *gorilla.Router
	}

	// Handle implements the mux.Engine interface from the krakend router package
	func (g gorillaEngine) Handle(pattern string, handler http.Handler) {
		g.r.Handle(pattern, handler)
	}

	// ServeHTTP implements the http:Handler interface from the stdlib
	func (g gorillaEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		g.r.ServeHTTP(w, r)
	}
}
