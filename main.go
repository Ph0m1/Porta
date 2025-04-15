package main

import (
	"flag"
	"github.com/ph0m1/p_gateway/config/viper"
	"github.com/ph0m1/p_gateway/logging/gologging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router/gin"
	"log"
	"os"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "etc/configuration.json", "Path to the configuration filename")
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

	logger := gologging.NewLogger(*logLevel, os.Stdout, "[X_X]")

	routerFactory := gin.DefaultFactory(proxy.DefaultFactory(logger), logger)

	routerFactory.New().Run(serviceConfig)
}
