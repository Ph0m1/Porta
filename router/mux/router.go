package mux

import (
	"fmt"
	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/logging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router"
	"net/http"
)

type factory struct {
	cfg Config
}

type Config struct {
	Engine         *http.ServeMux
	Middlewares    []HandlerMiddleware
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
}

type HandlerMiddleware interface {
	Handler(h http.Handler) http.Handler
}

func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{Config{Engine: http.NewServeMux(),
		Middlewares:    []HandlerMiddleware{},
		HandlerFactory: EndpointHandler,
		ProxyFactory:   pf,
		Logger:         logger,
	}}
}

func NewFactory(cfg Config) router.Factory {
	return factory{cfg}
}

func (rf factory) New() router.Router {
	return httpRouter{rf.cfg}
}

type httpRouter struct {
	cfg Config
}

func (r httpRouter) Run(cfg config.ServiceConfig) {
	if cfg.Debug {
		r.cfg.Engine.Handle("/__debug/", DebugHandler(r.cfg.Logger))
	}
	r.registerEndpoints(cfg.Endpoints)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r.handler(),
	}
	r.cfg.Logger.Critical(server.ListenAndServe())
}
func (r httpRouter) registerEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)

		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerEndpoint(c.Method, c.Endpoint, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r httpRouter) registerEndpoint(method, path string, handler http.HandlerFunc, toBackends int) {
	if method != "GET" && toBackends > 1 {
		r.cfg.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
		return
	}
	switch method {
	case "GET":
	case "POST":
	case "PUT":
	default:
		r.cfg.Logger.Error("Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug("registering the endpoint", method, path)
	r.cfg.Engine.Handle(path, handler)
}

func (r httpRouter) handler() http.Handler {
	var handler http.Handler
	handler = r.cfg.Engine
	for _, middleware := range r.cfg.Middlewares {
		r.cfg.Logger.Debug("Adding the middleware", middleware)
		handler = middleware.Handler(handler)
	}
	return handler
}
