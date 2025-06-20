package proxy

import (
	"github.com/ph0m1/porta/config"
	"github.com/ph0m1/porta/logging"
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
		backendProxy[i] = pf.newStack(backend)
	}
	p = NewMergeDataMiddleware(cfg)(backendProxy...)
	return
}

func (pf defaultFactory) newSingle(cfg *config.EndpointConfig) (p Proxy, err error) {
	return pf.newStack(cfg.Backend[0]), nil
}

func (pf defaultFactory) newStack(backend *config.Backend) (p Proxy) {
	p = pf.backendFactory(backend)
	p = NewRoundRobinLoadBalancedMiddleware(backend)(p)

	if backend.ConcurrentCalls > 1 {
		p = NewConcurrentMiddleware(backend)(p)
	}
	p = NewConcurrentMiddleware(backend)(p)
	return
}
