package proxy

import (
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
		p, err = pf.newSingle(cfg)
	default:
		p, err = pf.newMulti(cfg)
	}
	return
}

func (pf defaultFactory) newMulti(cfg *config.EndpointConfig) (p Proxy, err error) {
	backendProxy := make([]Proxy, len(cfg.Backend))

	for i, backend := range cfg.Backend {
		backendProxy[i] = pf.backendFactory(backend)
		backendProxy[i] = NewRoundRobinLoadBalancedMiddleware(backend)(backendProxy[i])
		if backend.ConcurrentCalls > 1 {
			backendProxy[i] = NewConcurrentMiddleware(backend)(backendProxy[i])
		}
		backendProxy[i] = NewRequestBuilderMiddleware(backend)(backendProxy[i])
	}
	p = NewMergeDataMiddleware(cfg)(backendProxy...)
	return
}

func (pf defaultFactory) newSingle(cfg *config.EndpointConfig) (p Proxy, err error) {
	p = pf.backendFactory(cfg.Backend[0])
	p = NewRoundRobinLoadBalancedMiddleware(cfg.Backend[0])(p)
	if cfg.Backend[0].ConcurrentCalls > 1 {
		p = NewConcurrentMiddleware(cfg.Backend[0])(p)
	}
	p = NewRequestBuilderMiddleware(cfg.Backend[0])(p)
	return
}
