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
	//engine := gin.Default()

	r.cfg.Engine.RedirectTrailingSlash = true
	r.cfg.Engine.RedirectFixedPath = true
	r.cfg.Engine.HandleMethodNotAllowed = true

	r.cfg.Engine.Use(r.cfg.Middlewares...)

	if cfg.Debug {
		r.registerDebugEndpoints()
	}
	r.registerEndpoints(cfg.Endpoints)

	r.cfg.Logger.Critical(r.cfg.Engine.Run(fmt.Sprintf(":%d", cfg.Port)))
}

func (r ginRouter) registerDebugEndpoints() {
	handler := DebugHandler(r.cfg.Logger)
	r.cfg.Engine.GET("/__debug/*param", handler)
	r.cfg.Engine.POST("/__debug/*param", handler)
	r.cfg.Engine.PUT("/__debug/*param", handler)
}
func (r ginRouter) registerEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}
		r.registerEndpoint(c.Method, c.Endpoint, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r ginRouter) registerEndpoint(method, path string, handler gin.HandlerFunc, toBackends int) {
	if method != "GET" && toBackends > 1 {
		r.cfg.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
		return
	}
	switch method {
	case "GET":
		r.cfg.Engine.GET(path, handler)
	case "POST":
		r.cfg.Engine.POST(path, handler)
	case "PUT":
		r.cfg.Engine.PUT(path, handler)
	case "PATCH":
		r.cfg.Engine.PATCH(path, handler)
	case "DELETE":
		r.cfg.Engine.DELETE(path, handler)

	default:
		r.cfg.Logger.Error("Unsupported method", method)
	}
}
