package proxy

import (
	"context"
	"github.com/ph0m1/p_gateway/config"
	"github.com/ph0m1/p_gateway/logging"
)

type Factory interface {
	New(cfg *config.EndpointConfig) (Proxy, error)
}

func DefaultFactory(logger logging.Logger) Factory {
	return NewDefaultFactory(httpProxy, logger)
}

func NewDefaultFactory(backendFactory BackendFactory, logger logging.Logger) Factory {
	return defaultFactory{backendFactory, logger}
}

type defaultFactory struct {
	backendFactory BackendFactory
	logger         logging.Logger
}

func (pf defaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	switch len(cfg.Backend) {
	case 0:
		err = ErrNoBackends
	case 1:

	}
	return
}

func (pf defaultFactory) newMulti(cfg *config.EndpointConfig) (p Proxy, err error) {
	backendProxy := make([]Proxy, len(cfg.Backend))

	for i, backend := range cfg.Backend {
		backendProxy[i] = pf.backendFactory(backend)
	}
	return
}

func (pf defaultFactory) newSingle() {

}
func (pf defaultFactory) newProxy() {

}
