package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/logging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router"
)

func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{pf, logger}
}

type factory struct {
	pf     proxy.Factory
	logger logging.Logger
}

func (rf factory) New() router.Router {
	return ginRouter{rf.pf, rf.logger}
}

type ginRouter struct {
	pf     proxy.Factory
	logger logging.Logger
}

func (r ginRouter) Run(cfg config.ServiceConfig) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		r.logger.Debug("Debug enabled")
	}
	engine := gin.Default()

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	for _, c := range cfg.Endpoints {
		proxyStack, err := r.pf.New(c)
		if err != nil {
			r.logger.Error("calling the ProxyFactory", err.Error())
			continue
		}
		handler := EndpointHandler(c, proxyStack)
		// add endpoint middleware components here:
		// logs, metrics, throttling, 3rd party integrations...
		// there are several in the package gin-gonic/gin and in the golang community

		switch c.Method {
		case "GET":
			engine.GET(c.Endpoint, handler)
		case "POST":
			if len(c.Backend) > 1 {
				r.logger.Error("POST endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			engine.POST(c.Endpoint, handler)
		case "PUT":
			if len(c.Backend) > 1 {
				r.logger.Error("PUT endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			engine.PUT(c.Endpoint, handler)
		default:
			r.logger.Error("Unsupported method", c.Method)
		}
	}
	if cfg.Debug {
		handler := DebugHandler(r.logger)
		engine.GET("/__debug/*param", handler)
		engine.POST("/__debug/*param", handler)
		engine.PUT("/__debug/*param", handler)
	}
	r.logger.Critical(engine.Run(fmt.Sprintf(":%d", cfg.Port)))
}
