package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"gopkg.in/unrolled/secure.v1"

	"github.com/ph0m1/porta/config/viper"
	"github.com/ph0m1/porta/logging/gologging"
	"github.com/ph0m1/porta/proxy"
	"github.com/ph0m1/porta/router/mux"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "configuration.json", "Path to configuration filename")
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
	logger, err := gologging.NewLogger(*logLevel, os.Stdout, "[O.o]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	secureMiddleware := secure.New(secure.Options{
		AllowedHosts:          []string{"127.0.0.1:8080", "example.com", "ssl.example.com"},
		SSLRedirect:           false,
		SSLHost:               "ssl.example.com",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            315360000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	})

	routerFactory := mux.NewFactory(mux.Config{
		Engine:         mux.DefaultEngine(),
		ProxyFactory:   proxy.DefaultFactory(logger),
		Middlewares:    []mux.HandlerMiddleware{secureMiddleware},
		Logger:         logger,
		HandlerFactory: mux.EndpointHandler,
	})

	routerFactory.New().Run(serviceConfig)
}
