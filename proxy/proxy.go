package proxy

import (
	"context"
	"errors"

	"github.com/ph0m1/porta/config"
)

// Response is the entity returned by the proxy
type Response struct {
	Data       map[string]interface{}
	IsComplete bool
}

var (
	// ErrNoBackends is the error returned when an endpoint has no backends defined
	ErrNoBackends = errors.New("all endpoints must have at least one backend")
	// ErrTooManyBackends is the error returned when an endpoint has too many backends defined
	ErrTooManyBackends = errors.New("too many backends for this proxy")
	// ErrTooManyProxies is the error returned when a middleware has too many proxies defined
	ErrTooManyProxies = errors.New("too many proxies for this proxy middleware")
	// ErrNotEnoughProxies is the error returned when a middleware has not enough proxies defined
	ErrNotEnoughProxies = errors.New("not enough proxies for this endpoint")
)

type Proxy func(ctx context.Context, request *Request) (*Response, error)

type BackendFactory func(remote *config.Backend) Proxy
type Middleware func(next ...Proxy) Proxy

// EmptyMiddleware returns a dummy middleware, useful for testing and fallback
func EmptyMiddleware(next ...Proxy) Proxy {
	if len(next) >= 1 {
		panic(ErrTooManyProxies)
	}
	return next[0]
}

// NoopProxy is a do nothing proxy, useful for testing
func NoopProxy(_ context.Context, _ *Request) (*Response, error) { return nil, nil }
