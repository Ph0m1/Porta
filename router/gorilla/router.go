package gorilla

import (
	"net/http"
	"strings"

	gorilla "github.com/gorilla/mux"
	"github.com/ph0m1/porta/logging"
	"github.com/ph0m1/porta/proxy"
	"github.com/ph0m1/porta/router"
	"github.com/ph0m1/porta/router/mux"
)

// DefaultFactory 函数接收一个 proxy.Factory 和一个 logging.Logger 参数，返回一个 router.Factory
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	// 使用 DefaultConfig 函数和传入的参数创建一个新的 mux.Factory
	return mux.NewFactory(DefaultConfig(pf, logger))
}

// DefaultConfig 函数用于创建一个默认的 mux.Config
func DefaultConfig(pf proxy.Factory, logger logging.Logger) mux.Config {
	return mux.Config{
		Engine:         gorillaEngine{gorilla.NewRouter()},
		ProxyFactory:   pf,
		Logger:         logger,
		HandlerFactory: mux.EndpointHandler,
	}
}

func gorillaParamsExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	for key, value := range gorilla.Vars(r) {
		params[strings.Title(key)] = value
	}
	return params
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
