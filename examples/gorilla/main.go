package main

import (
	"flag"
	"log"

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
}
