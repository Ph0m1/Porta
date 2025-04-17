package mux

import (
	"fmt"
	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/logging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router"
	"net/http"
)

func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{pf, logger}
}

type factory struct {
	pf     proxy.Factory
	logger logging.Logger
}

func (rf factory) New() router.Router {
	return httpRouter{rf.pf, rf.logger}
}

type httpRouter struct {
	pf     proxy.Factory
	logger logging.Logger
}

func (r httpRouter) Run(cfg config.ServiceConfig) {
	mux := http.NewServeMux()
	for _, c := range cfg.Endpoints {
		proxyStack, err := r.pf.New(c)
		if err != nil {
			r.logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		switch c.Method {
		case "GET":
		case "POST":
			if len(c.Backend) > 1 {
				r.logger.Error("POST endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
		case "PUT":
			if len(c.Backend) > 1 {
				r.logger.Error("PUT endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
		default:
			r.logger.Error("Unsupported method", c.Method)
			continue
		}
		mux.Handle(c.Endpoint, EndpointHandler(c, proxyStack))
	}

	if cfg.Debug {
		mux.Handle("/__debug/", DebugHandler(r.logger))
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	r.logger.Critical(server.ListenAndServe())

}
