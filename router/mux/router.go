package mux

import (
	"fmt"
	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/logging"
	"github.com/ph0m1/porta/proxy"
	"github.com/ph0m1/porta/router"
	"net/http"
)

const DefaultDebugPattern = "/__debug/"

// Engine defines the minimum required interface for the mux compatible engine
type Engine interface {
	http.Handler
	Handle(pattern string, handler http.Handler)
}

// DefaultEngine returns a new engine using the http.ServeMux router
func DefaultEngine() *http.ServeMux {
	return http.NewServeMux()
}

type factory struct {
	cfg Config
}

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         Engine
	Middlewares    []HandlerMiddleware
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
	DebugPattern   string
}

// HandlerMiddleware is the interface for rhe decorators over the http.Handler
type HandlerMiddleware interface {
	Handler(h http.Handler) http.Handler
}

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{Config{
		Engine:         DefaultEngine(),
		Middlewares:    []HandlerMiddleware{},
		HandlerFactory: EndpointHandler,
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   DefaultDebugPattern,
	}}
}

// NewFactory returns a net/http mux router factory with the injected configuration
func NewFactory(cfg Config) router.Factory {
	if cfg.DebugPattern == "" {
		cfg.DebugPattern = DefaultDebugPattern
	}
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
		r.cfg.Engine.Handle(r.cfg.DebugPattern, DebugHandler(r.cfg.Logger))
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
