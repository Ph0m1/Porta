package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/logging"
	"github.com/ph0m1/p_gateway/proxy"
	"github.com/ph0m1/p_gateway/router"
)

type Config struct {
	Engine         *gin.Engine
	Middlewares    []gin.HandlerFunc
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
}

func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{
		Config{
			Engine:         gin.Default(),
			Middlewares:    []gin.HandlerFunc{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
		},
	}
}

func NewFactory(cfg Config) router.Factory {
	return factory{cfg}
}

type factory struct {
	cfg Config
}

func (rf factory) New() router.Router {
	return ginRouter{rf.cfg}
}

type ginRouter struct {
	cfg Config
}

func (r ginRouter) Run(cfg config.ServiceConfig) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		r.cfg.Logger.Debug("Debug enabled")
	}
	engine := gin.Default()

	r.cfg.Engine.RedirectTrailingSlash = true
	r.cfg.Engine.RedirectFixedPath = true
	r.cfg.Engine.HandleMethodNotAllowed = true

	r.cfg.Engine.Use(r.cfg.Middlewares...)

	for _, c := range cfg.Endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}
		handler := r.cfg.HandlerFactory(c, proxyStack)
		// add endpoint middleware components here:
		// logs, metrics, throttling, 3rd party integrations...
		// there are several in the package gin-gonic/gin and in the golang community

		switch c.Method {
		case "GET":
			r.cfg.Engine.GET(c.Endpoint, handler)
		case "POST":
			if len(c.Backend) > 1 {
				r.cfg.Logger.Error("POST endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			engine.POST(c.Endpoint, handler)
		case "PUT":
			if len(c.Backend) > 1 {
				r.cfg.Logger.Error("PUT endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			r.cfg.Engine.PUT(c.Endpoint, handler)
		default:
			r.cfg.Logger.Error("Unsupported method", c.Method)
		}
	}
	if cfg.Debug {
		handler := DebugHandler(r.cfg.Logger)
		r.cfg.Engine.GET("/__debug/*param", handler)
		r.cfg.Engine.POST("/__debug/*param", handler)
		r.cfg.Engine.PUT("/__debug/*param", handler)
	}
	r.cfg.Logger.Critical(engine.Run(fmt.Sprintf(":%d", cfg.Port)))
}
